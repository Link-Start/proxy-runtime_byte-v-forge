package app

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/proto"
)

const chainHopEnrichmentTimeout = 2 * time.Second

func (r *Runtime) routePlanHops(ctx context.Context, line *scoredLineCandidate, gateway scoredGatewayCandidate) []*proxyruntimev1.EgressHop {
	hops := make([]*proxyruntimev1.EgressHop, 0, 2)
	if line != nil && line.proto != nil {
		hop := &proxyruntimev1.EgressHop{
			HopId: "line:" + line.proto.GetSourceId() + ":" + line.proto.GetNodeId(),
			Order: uint32(len(hops) + 1),
			Role:  proxyruntimev1.EgressHopRole_EGRESS_HOP_ROLE_FORWARD,
			Selector: &proxyruntimev1.ProxySelectorPolicy{
				Strategy: proxyruntimev1.ProxySelectorStrategy_PROXY_SELECTOR_STRATEGY_FIFO,
			},
			Endpoints: []*proxyruntimev1.ProxyEndpoint{{
				Id:           line.proto.GetSourceId() + ":" + line.proto.GetNodeId(),
				ProviderId:   sourceRuntimeProviderID,
				UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_SIMPLE_PROXY,
				RotationMode: proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_NONE,
				Labels: map[string]string{
					"source_kind":         line.proto.GetSourceKind().String(),
					"source_id":           line.proto.GetSourceId(),
					"source_display_name": line.proto.GetSourceDisplayName(),
					"node_id":             line.proto.GetNodeId(),
					"node_display_name":   line.proto.GetDisplayName(),
					"status":              line.proto.GetStatus().String(),
					"delay_ms":            fmtUint32(line.proto.GetDelayMs()),
				},
			}},
		}
		hops = append(hops, r.enrichRouteHop(ctx, hop, func(resolveCtx context.Context) (string, error) {
			return r.sourcePlane.ResolveNodePublicIP(resolveCtx, line.proto.GetSourceId(), line.proto.GetNodeId(), line.proto.GetDisplayName())
		}))
	}
	if gateway.proto != nil {
		hop := &proxyruntimev1.EgressHop{
			HopId: "dynamic-gateway:" + gateway.proto.GetProviderAccountId() + ":" + gateway.proto.GetGatewayId(),
			Order: uint32(len(hops) + 1),
			Role:  proxyruntimev1.EgressHopRole_EGRESS_HOP_ROLE_EXIT,
			Selector: &proxyruntimev1.ProxySelectorPolicy{
				Strategy: proxyruntimev1.ProxySelectorStrategy_PROXY_SELECTOR_STRATEGY_FIFO,
			},
			Endpoints: []*proxyruntimev1.ProxyEndpoint{{
				Id:           gateway.proto.GetProviderAccountId() + ":" + gateway.proto.GetGatewayId(),
				ProviderId:   gateway.proto.GetProviderId(),
				UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP,
				RotationMode: proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION,
				Labels: map[string]string{
					"source_kind":          proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_DYNAMIC_IP.String(),
					"provider_account_id":  gateway.proto.GetProviderAccountId(),
					"gateway_id":           gateway.proto.GetGatewayId(),
					"gateway_display_name": gateway.proto.GetDisplayName(),
				},
			}},
		}
		hops = append(hops, r.enrichRouteHop(ctx, hop, func(resolveCtx context.Context) (string, error) {
			return resolvePublicIP(resolveCtx, networkAddressHost(gateway.gateway.Addr))
		}))
	}
	return hops
}

func (r *Runtime) enrichRouteHop(ctx context.Context, hop *proxyruntimev1.EgressHop, resolve func(context.Context) (string, error)) *proxyruntimev1.EgressHop {
	if hop == nil || resolve == nil {
		return hop
	}
	enrichCtx, cancel := context.WithTimeout(ctx, chainHopEnrichmentTimeout)
	defer cancel()
	results := make(chan *proxyruntimev1.EgressHop, 1)
	go func() {
		enriched := proto.Clone(hop).(*proxyruntimev1.EgressHop)
		ip, err := resolve(enrichCtx)
		if err != nil {
			r.logger.Warn("resolve egress route hop public ip failed", "hop_id", hop.GetHopId(), "error", err)
			results <- enriched
			return
		}
		r.fillRouteHopGeo(enrichCtx, enriched, ip)
		results <- enriched
	}()
	select {
	case <-ctx.Done():
		return hop
	case <-enrichCtx.Done():
		return hop
	case enriched := <-results:
		return enriched
	}
}

func (r *Runtime) fillRouteHopGeo(ctx context.Context, hop *proxyruntimev1.EgressHop, ip string) {
	if hop == nil {
		return
	}
	ip = strings.TrimSpace(ip)
	if net.ParseIP(ip) == nil {
		return
	}
	setRouteHopLabel(hop, "observed_ip", ip)
	geo, err := r.lookupIPGeo(ctx, ip)
	if err != nil {
		r.logger.Warn("resolve egress route hop geo failed", "hop_id", hop.GetHopId(), "observed_ip", ip, "error", err)
		return
	}
	setRouteHopLabel(hop, "country_code", geo.CountryCode)
	setRouteHopLabel(hop, "region", geo.Region)
	setRouteHopLabel(hop, "city", geo.City)
}

func routeHopByRole(plan *proxyruntimev1.EgressRoutePlan, role proxyruntimev1.EgressHopRole) *proxyruntimev1.EgressHop {
	for _, hop := range plan.GetRoute().GetHops() {
		if hop.GetRole() == role {
			return hop
		}
	}
	return nil
}

func routeHopLabel(hop *proxyruntimev1.EgressHop, key string) string {
	for _, endpoint := range hop.GetEndpoints() {
		if value := strings.TrimSpace(endpoint.GetLabels()[key]); value != "" {
			return value
		}
	}
	return ""
}

func setRouteHopLabel(hop *proxyruntimev1.EgressHop, key string, value string) {
	if hop == nil || key == "" || strings.TrimSpace(value) == "" {
		return
	}
	endpoints := hop.GetEndpoints()
	if len(endpoints) == 0 {
		endpoints = []*proxyruntimev1.ProxyEndpoint{{}}
		hop.Endpoints = endpoints
	}
	if endpoints[0].Labels == nil {
		endpoints[0].Labels = map[string]string{}
	}
	endpoints[0].Labels[key] = strings.TrimSpace(value)
}

func fmtUint32(value uint32) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(value), 10)
}

func networkAddressHost(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return strings.Trim(host, "[]")
	}
	if parsed, err := url.Parse(addr); err == nil && parsed != nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return strings.Trim(addr, "[]")
}

func resolvePublicIP(ctx context.Context, host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ip := addr.IP.To4(); ip != nil {
			return ip.String(), nil
		}
	}
	if len(addrs) > 0 {
		return addrs[0].IP.String(), nil
	}
	return "", nil
}
