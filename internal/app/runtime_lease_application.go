package app

import (
	"context"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type runtimeLeaseApplication struct {
	list   *leaseapp.Application
	leases leaseCoordinator
}

func newRuntimeLeaseApplication(runtime *Runtime) runtimeLeaseApplication {
	return runtimeLeaseApplication{list: leaseapp.NewApplication(runtime.store), leases: runtime.leaseCoordinator}
}

func (s *RuntimeService) ListProxyDynamicLeases(ctx context.Context, _ *proxyruntimev1.ListProxyDynamicLeasesRequest) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.listProxyDynamicLeases(ctx, leaseapp.DefaultListOptions())
}

func (s *RuntimeService) AcquireProxyLease(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.acquireProxyLease(ctx, nil, req)
}

func (s *RuntimeService) acquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.leases.AcquireProxyLease(ctx, httpReq, req)
}

func (s *RuntimeService) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	return s.leases.ReleaseProxyLease(ctx, req)
}

func (s *RuntimeService) listProxyDynamicLeases(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.leases.ListProxyDynamicLeaseFacts(ctx, options)
}

func (a runtimeLeaseApplication) ListProxyDynamicLeaseFacts(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	leases, err := a.list.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyDynamicLeasesResponse{Leases: leases}, nil
}

func (a runtimeLeaseApplication) AcquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	lease, err := a.leases.acquireLease(ctx, httpReq, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.AcquireProxyLeaseResponse{Lease: lease, Egress: lease.GetEgress(), SelectionPlan: lease.GetSelectionPlan()}, nil
}

func (a runtimeLeaseApplication) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	lease, err := a.leases.releaseLease(ctx, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ReleaseProxyLeaseResponse{Lease: lease}, nil
}
