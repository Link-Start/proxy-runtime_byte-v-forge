package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/geox"
	"github.com/byte-v-forge/common-lib/randx"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultResolveProxyTTL = 10 * time.Minute

type runtimeProxyApplication struct {
	runtime *Runtime
}

func newRuntimeProxyApplication(runtime *Runtime) runtimeProxyApplication {
	return runtimeProxyApplication{runtime: runtime}
}

func (s *RuntimeService) ResolveProxy(ctx context.Context, req *proxyruntimev1.ResolveProxyRequest) (*proxyruntimev1.ResolveProxyResponse, error) {
	return s.resolveProxy(ctx, nil, req)
}

func (s *RuntimeService) resolveProxy(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.ResolveProxyRequest) (*proxyruntimev1.ResolveProxyResponse, error) {
	return s.proxies.ResolveProxy(ctx, httpReq, req)
}

func (a runtimeProxyApplication) ResolveProxy(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.ResolveProxyRequest) (*proxyruntimev1.ResolveProxyResponse, error) {
	if req == nil {
		return nil, invalidArgument("resolve proxy request is required", nil)
	}
	switch req.GetProxyKind() {
	case proxyruntimev1.ProxyKind_PROXY_KIND_DYNAMIC_IP:
		return a.resolveDynamicProxy(ctx, httpReq, req)
	case proxyruntimev1.ProxyKind_PROXY_KIND_STATIC_IP:
		return nil, unavailable("static proxies are managed by Mihomo dashboard", nil)
	case proxyruntimev1.ProxyKind_PROXY_KIND_SUBSCRIPTION:
		return nil, unavailable("subscription proxies are managed by Mihomo dashboard", nil)
	case proxyruntimev1.ProxyKind_PROXY_KIND_FIXED_PROXY:
		return nil, unavailable("fixed proxies are managed by Mihomo dashboard", nil)
	default:
		return nil, invalidArgument("proxy_kind is required", nil)
	}
}

func (a runtimeProxyApplication) resolveDynamicProxy(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.ResolveProxyRequest) (*proxyruntimev1.ResolveProxyResponse, error) {
	runtime := a.runtime
	policy := resolveProxyRoutePolicy(req)
	policy.RequireDynamicExit = true
	policy.AllowDirectDynamicGateway = false
	ttl := resolveProxyTTL(req)
	assignmentID, err := dynamicProxyAssignmentID(req, policy)
	if err != nil {
		return nil, internalError("create dynamic proxy assignment failed", err)
	}
	lease, err := runtime.leaseCoordinator.acquireLease(ctx, httpReq, &proxyruntimev1.AcquireProxyLeaseRequest{
		AccountId: assignmentID,
		Purpose:   policy.GetPurpose(),
		ForceNew:  true,
		Policy: &proxyruntimev1.ProxySessionPolicy{
			Mode:         proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY,
			Region:       firstNonEmpty(policy.GetCountryCode(), policy.GetRegion()),
			StickyTtl:    durationpb.New(ttl),
			UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP,
			RotationMode: proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION,
			Labels: map[string]string{
				"country_code":   policy.GetCountryCode(),
				"region":         policy.GetRegion(),
				"purpose":        policy.GetPurpose(),
				"selection_seed": assignmentID,
			},
		},
		RoutePolicy: policy,
	})
	if err != nil {
		return nil, unavailable("resolve dynamic proxy failed", err)
	}
	proxyURL, err := proxyURLFromEndpoint(lease.GetEgress())
	if err != nil {
		return nil, unavailable("dynamic proxy endpoint is unavailable", err)
	}
	resource := &proxyruntimev1.ProxyResourceRef{
		ProxyKind:   proxyruntimev1.ProxyKind_PROXY_KIND_DYNAMIC_IP,
		SourceId:    lease.GetProviderAccountId(),
		NodeId:      lease.GetSession().GetSessionId(),
		DisplayName: "Dynamic IP",
		SourceKind:  proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_DYNAMIC_IP,
		RegionCodes: cleanRegionCodes([]string{
			policy.GetCountryCode(),
			policy.GetRegion(),
		}),
	}
	return &proxyruntimev1.ResolveProxyResponse{Proxy: &proxyruntimev1.ResolvedProxy{
		ProxyUrl:     proxyURL,
		Endpoint:     lease.GetEgress(),
		ProxyKind:    proxyruntimev1.ProxyKind_PROXY_KIND_DYNAMIC_IP,
		Resource:     resource,
		AssignmentId: lease.GetAccountId(),
		LeaseId:      lease.GetLeaseId(),
		ExpiresAt:    lease.GetExpiresAt(),
		RoutePlan:    lease.GetRoutePlan(),
	}, Candidates: []*proxyruntimev1.ProxyResourceRef{resource}}, nil
}

func resolveProxyRoutePolicy(req *proxyruntimev1.ResolveProxyRequest) *proxyruntimev1.EgressRoutePolicy {
	policy := &proxyruntimev1.EgressRoutePolicy{
		CountryCode: geox.NormalizeCountryAlpha2(req.GetCountryCode()),
		Region:      strings.ToUpper(strings.TrimSpace(req.GetRegion())),
		Purpose:     firstNonEmpty(req.GetPurpose(), "PROXY_RESOLVE"),
		Strategy:    req.GetStrategy(),
		MaxAttempts: 10,
	}
	if policy.Strategy == proxyruntimev1.ProxySelectorStrategy_PROXY_SELECTOR_STRATEGY_UNSPECIFIED {
		policy.Strategy = proxyruntimev1.ProxySelectorStrategy_PROXY_SELECTOR_STRATEGY_HASH_TARGET_HOST
	}
	return policy
}

func resolveProxySelectionKey(req *proxyruntimev1.ResolveProxyRequest, policy *proxyruntimev1.EgressRoutePolicy) string {
	return firstNonEmpty(
		req.GetStickinessKey(),
		strings.Join([]string{req.GetProxyKind().String(), policy.GetCountryCode(), policy.GetRegion(), policy.GetPurpose(), req.GetTargetHost()}, ":"),
		req.GetProxyKind().String(),
	)
}

func dynamicProxyAssignmentID(req *proxyruntimev1.ResolveProxyRequest, policy *proxyruntimev1.EgressRoutePolicy) (string, error) {
	token, err := randx.Hex(8)
	if err != nil {
		return "", err
	}
	key := strings.Join([]string{
		req.GetProxyKind().String(),
		policy.GetCountryCode(),
		policy.GetRegion(),
		policy.GetPurpose(),
		token,
	}, ":")
	return "dynamic-" + shortHash(key), nil
}

func resolveProxyTTL(req *proxyruntimev1.ResolveProxyRequest) time.Duration {
	if req.GetTtl() == nil || req.GetTtl().AsDuration() <= 0 {
		return defaultResolveProxyTTL
	}
	return req.GetTtl().AsDuration()
}

func (r *Runtime) observeProxyURL(ctx context.Context, key string, proxyURL *url.URL, status proxyruntimev1.ProxySourceNodeStatus, delayMS uint32) *proxyruntimev1.ProxyNodeObservation {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if observation, ok, err := r.nodeObservations.Load(ctx, key); err == nil && ok {
		return observation
	}
	return r.observeProxyURLFresh(ctx, key, proxyURL, status, delayMS)
}

func (r *Runtime) observeProxyURLFresh(ctx context.Context, key string, proxyURL *url.URL, status proxyruntimev1.ProxySourceNodeStatus, delayMS uint32) *proxyruntimev1.ProxyNodeObservation {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	now := time.Now().UTC()
	observation := &proxyruntimev1.ProxyNodeObservation{
		Status:     firstObservedStatus(status),
		DelayMs:    delayMS,
		ObservedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(proxyNodeObservationTTL)),
	}
	if proxyURL == nil || strings.TrimSpace(proxyURL.Host) == "" {
		observation.Status = proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNAVAILABLE
		observation.ErrorMessage = "proxy endpoint is unavailable"
		_ = r.nodeObservations.Save(ctx, key, observation)
		return observation
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		observation.ErrorMessage = "proxy observation settings unavailable"
		_ = r.nodeObservations.Save(ctx, key, observation)
		return observation
	}
	timeout := proxyExitIPTimeout(settings)
	client, err := runtimehttp.NewWithProxy(timeout, proxyURL.String(), runtimehttp.CommonProxySchemes...)
	if err != nil {
		observation.Status = proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNAVAILABLE
		observation.ErrorMessage = "proxy observation client unavailable"
		_ = r.nodeObservations.Save(ctx, key, observation)
		return observation
	}
	started := time.Now()
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	exitIP, err := r.probeExitIP(probeCtx, client)
	cancel()
	if observation.DelayMs == 0 {
		observation.DelayMs = uint32(time.Since(started).Milliseconds())
	}
	if err != nil {
		observation.Status = proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNAVAILABLE
		observation.ErrorMessage = "proxy exit IP probe failed"
		_ = r.nodeObservations.Save(ctx, key, observation)
		return observation
	}
	observation.Status = proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_AVAILABLE
	observation.ExitIp = exitIP
	if geo, err := r.lookupIPGeo(ctx, exitIP); err == nil {
		observation.ExitGeo = &proxyruntimev1.ProxyExitGeo{Ip: exitIP, CountryCode: geo.CountryCode, Region: geo.Region, City: geo.City, CheckedAt: timestamppb.Now()}
	}
	if fraudCheck, err := r.checkIPFraud(ctx, exitIP, settings); err == nil {
		observation.IpFraudCheck = fraudCheck
		if proxyObservationBlocked(observation) {
			observation.Status = proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNAVAILABLE
		}
	}
	_ = r.nodeObservations.Save(ctx, key, observation)
	return observation
}

func firstObservedStatus(status proxyruntimev1.ProxySourceNodeStatus) proxyruntimev1.ProxySourceNodeStatus {
	if status == proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNSPECIFIED {
		return proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNKNOWN
	}
	return status
}

func proxyObservationBlocked(observation *proxyruntimev1.ProxyNodeObservation) bool {
	switch observation.GetIpFraudCheck().GetRiskLevel() {
	case proxyruntimev1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_HIGH,
		proxyruntimev1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_CRITICAL:
		return true
	default:
		return false
	}
}

func proxyURLFromEndpoint(endpoint *proxyruntimev1.ProxyEndpoint) (string, error) {
	if endpoint == nil || endpoint.GetPort() == 0 {
		return "", fmt.Errorf("proxy endpoint is unavailable")
	}
	host := strings.TrimSpace(endpoint.GetHost())
	if localOnlyHost(host) {
		host = firstLocalAdvertisedIP()
	}
	if host == "" {
		return "", fmt.Errorf("proxy endpoint host is unavailable")
	}
	proxyURL := &url.URL{Scheme: protocolName(endpoint.GetProtocol()), Host: net.JoinHostPort(host, fmt.Sprintf("%d", endpoint.GetPort()))}
	username := strings.TrimSpace(endpoint.GetLabels()["proxy_username"])
	password := strings.TrimSpace(endpoint.GetLabels()["proxy_password"])
	if username != "" || password != "" {
		proxyURL.User = url.UserPassword(username, password)
	}
	return proxyURL.String(), nil
}
