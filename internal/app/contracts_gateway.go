package app

import (
	"context"
	"fmt"
	"net"
	"os"

	newpb "github.com/byte-v-forge/contracts/byte/v/forge/contracts/proxygateway/v1"
	oldpb "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	checkapp "github.com/byte-v-forge/proxy-gateway/internal/app/proxycheck/application"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type contractsGatewayServer struct {
	newpb.UnimplementedProxyGatewayServiceServer
	leases         runtimeLeaseApplication
	checks         checkapp.Service
	status         runtimeStatusApplication
	advertisedHost string
}

func newContractsGatewayServer(r *Runtime) *contractsGatewayServer {
	return &contractsGatewayServer{
		leases:         r.leases,
		checks:         checkapp.New(runtimeCheckDependencies(r)),
		status:         newRuntimeStatusApplication(runtimeStatusDependencies(r)),
		advertisedHost: r.cfg.SessionListener.AdvertisedHost,
	}
}

func (r *Runtime) serveGRPC(ctx context.Context, errCh chan<- error) {
	addr := os.Getenv("PROXY_GATEWAY_GRPC_ADDR")
	if addr == "" {
		addr = ":9090"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		errCh <- fmt.Errorf("serve proxy-gateway grpc: listen %s: %w", addr, err)
		return
	}
	server := grpc.NewServer()
	newpb.RegisterProxyGatewayServiceServer(server, newContractsGatewayServer(r))
	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
	r.logger.Info("proxy-gateway grpc listening", "addr", addr)
	if err := server.Serve(listener); err != nil {
		errCh <- fmt.Errorf("serve proxy-gateway grpc: %w", err)
	}
}

func (s *contractsGatewayServer) AcquireLease(ctx context.Context, req *newpb.AcquireLeaseRequest) (*newpb.AcquireLeaseResponse, error) {
	old := &oldpb.AcquireProxyLeaseRequest{
		Purpose: req.GetPurpose(),
		Policy:  newGeoSessionToOldSessionPolicy(req.GetGeo(), req.GetSession()),
	}
	resp, err := s.leases.AcquireProxyLease(ctx, s.advertisedHost, old)
	if err != nil {
		return nil, err
	}
	return &newpb.AcquireLeaseResponse{Lease: oldLeaseToNew(resp.GetLease())}, nil
}

func (s *contractsGatewayServer) ReleaseLease(ctx context.Context, req *newpb.ReleaseLeaseRequest) (*newpb.ReleaseLeaseResponse, error) {
	old := &oldpb.ReleaseProxyLeaseRequest{LeaseId: req.GetUid()}
	resp, err := s.leases.ReleaseProxyLease(ctx, old)
	if err != nil {
		return nil, err
	}
	return &newpb.ReleaseLeaseResponse{Lease: oldLeaseToNew(resp.GetLease())}, nil
}

func (s *contractsGatewayServer) GetLease(ctx context.Context, req *newpb.GetLeaseRequest) (*newpb.GetLeaseResponse, error) {
	old, err := s.leases.GetProxyDynamicLeaseFact(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &newpb.GetLeaseResponse{Lease: oldLeaseToNew(old)}, nil
}

func (s *contractsGatewayServer) GetLeaseCredentials(ctx context.Context, req *newpb.GetLeaseCredentialsRequest) (*newpb.GetLeaseCredentialsResponse, error) {
	return nil, grpcstatus.Error(codes.Unimplemented, "GetLeaseCredentials: gateway authenticates via session-bound endpoint; per-lease credentials are not yet exposed")
}

func (s *contractsGatewayServer) GetExitCheck(ctx context.Context, req *newpb.GetExitCheckRequest) (*newpb.GetExitCheckResponse, error) {
	old, err := s.checks.GetProxyExitCheckSnapshot(ctx, &oldpb.GetProxyExitCheckSnapshotRequest{ListenerId: req.GetUid()})
	if err != nil {
		return nil, err
	}
	return &newpb.GetExitCheckResponse{Snapshot: oldExitCheckToNew(old.GetSnapshot())}, nil
}

func (s *contractsGatewayServer) GetGatewayStatus(ctx context.Context, req *newpb.GetGatewayStatusRequest) (*newpb.GetGatewayStatusResponse, error) {
	resp, err := s.status.GetProxyGatewayStatus(ctx)
	if err != nil {
		return nil, err
	}
	rs := resp.GetStatus()
	return &newpb.GetGatewayStatusResponse{
		Status: &newpb.GatewayStatus{
			Ready:   rs.GetReady(),
			Version: "v1",
		},
	}, nil
}

// newGeoSessionToOldSessionPolicy collapses the new GeoTarget + SessionPolicy
// pair into the legacy ProxySessionPolicy. Dimensions the legacy schema does
// not carry (city / postal_code / asn / killswitch / business_identity) are
// dropped here until the underlying business layer is upgraded.
func newGeoSessionToOldSessionPolicy(geo *newpb.GeoTarget, sess *newpb.SessionPolicy) *oldpb.ProxySessionPolicy {
	if geo == nil && sess == nil {
		return nil
	}
	out := &oldpb.ProxySessionPolicy{
		Mode: oldpb.ProxySessionMode_PROXY_SESSION_MODE_ROTATING,
	}
	if geo != nil {
		// legacy Region absorbs whichever geo dimension is most specific.
		switch {
		case geo.GetState() != "":
			out.Region = geo.GetState()
		case geo.GetCountryCode() != "":
			out.Region = geo.GetCountryCode()
		}
	}
	if sess != nil {
		if sess.GetRotation() == newpb.SessionRotation_SESSION_ROTATION_STICKY || sess.GetStickyTtl().AsDuration() > 0 {
			out.Mode = oldpb.ProxySessionMode_PROXY_SESSION_MODE_STICKY
			out.StickyTtl = sess.GetStickyTtl()
		}
	}
	return out
}

func oldLeaseToNew(l *oldpb.ProxyDynamicLease) *newpb.ProxyLease {
	if l == nil {
		return nil
	}
	return &newpb.ProxyLease{
		Uid:        l.GetLeaseId(),
		Purpose:    l.GetPurpose(),
		Status:     oldLeaseStatusToNew(l.GetStatus()),
		Endpoint:   oldEndpointToNew(l.GetEgress()),
		CreateTime: l.GetAcquiredAt(),
		ExpireTime: l.GetExpiresAt(),
		Error:      errorMessageToStatus(l.GetErrorMessage()),
	}
}

func errorMessageToStatus(msg string) *statuspb.Status {
	if msg == "" {
		return nil
	}
	return &statuspb.Status{Code: int32(codes.Unknown), Message: msg}
}

func oldEndpointToNew(e *oldpb.ProxyEndpoint) *newpb.ProxyEndpoint {
	if e == nil {
		return nil
	}
	return &newpb.ProxyEndpoint{
		Protocol: oldProtocolToNew(e.GetProtocol()),
		Host:     e.GetHost(),
		Port:     e.GetPort(),
	}
}

func oldLeaseStatusToNew(s oldpb.ProxyDynamicLeaseStatus) newpb.LeaseStatus {
	switch s {
	case oldpb.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE:
		return newpb.LeaseStatus_LEASE_STATUS_ACTIVE
	case oldpb.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED:
		return newpb.LeaseStatus_LEASE_STATUS_EXPIRED
	case oldpb.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED:
		return newpb.LeaseStatus_LEASE_STATUS_RELEASED
	case oldpb.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED:
		return newpb.LeaseStatus_LEASE_STATUS_FAILED
	default:
		return newpb.LeaseStatus_LEASE_STATUS_UNSPECIFIED
	}
}

func oldProtocolToNew(p oldpb.ProxyProtocol) newpb.ProxyProtocol {
	switch p {
	case oldpb.ProxyProtocol_PROXY_PROTOCOL_HTTP:
		return newpb.ProxyProtocol_PROXY_PROTOCOL_HTTP
	case oldpb.ProxyProtocol_PROXY_PROTOCOL_SOCKS5:
		return newpb.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	default:
		return newpb.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
}

func oldExitCheckToNew(s *oldpb.ProxyExitCheckSnapshot) *newpb.ExitCheckSnapshot {
	if s == nil {
		return nil
	}
	return &newpb.ExitCheckSnapshot{
		LeaseUid: s.GetListenerId(),
		Exit: &newpb.ExitInfo{
			Ip:          s.GetProxyExitIp().GetIp(),
			CountryCode: s.GetProxyExitGeo().GetCountryCode(),
			State:       s.GetProxyExitGeo().GetRegion(),
			City:        s.GetProxyExitGeo().GetCity(),
		},
		Risk:       oldEdgeAccessToRisk(s.GetEdgeAccessCheck()),
		UpdateTime: s.GetUpdatedAt(),
	}
}

func oldEdgeAccessToRisk(c *oldpb.ProxyEdgeAccessCheck) *newpb.ExitRisk {
	if c == nil {
		return nil
	}
	return &newpb.ExitRisk{
		Level:     mapEdgeRiskLevel(c.GetRiskLevel()),
		Score:     c.GetRiskScore(),
		CheckTime: c.GetCheckedAt(),
	}
}

func mapEdgeRiskLevel(l oldpb.ProxyEdgeAccessRiskLevel) newpb.ExitRiskLevel {
	switch l {
	case oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW:
		return newpb.ExitRiskLevel_EXIT_RISK_LEVEL_LOW
	case oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_MEDIUM:
		return newpb.ExitRiskLevel_EXIT_RISK_LEVEL_MEDIUM
	case oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH,
		oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_BLOCK_LIKELY,
		oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_CHALLENGE_LIKELY:
		return newpb.ExitRiskLevel_EXIT_RISK_LEVEL_HIGH
	case oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN,
		oldpb.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNSUPPORTED:
		return newpb.ExitRiskLevel_EXIT_RISK_LEVEL_UNKNOWN
	default:
		return newpb.ExitRiskLevel_EXIT_RISK_LEVEL_UNSPECIFIED
	}
}
