package app

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/proto"
)

const routeHopEnrichmentTimeout = 2 * time.Second

func (p *dynamicGatewaySelector) dynamicGatewayRouteHops(ctx context.Context, gateway scoredGatewayCandidate) []*proxyruntimev1.EgressHop {
	hops := make([]*proxyruntimev1.EgressHop, 0, 1)
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
		hops = append(hops, p.enrichRouteHop(ctx, hop, func(resolveCtx context.Context) (string, error) {
			return resolvePublicIP(resolveCtx, networkAddressHost(gateway.gateway.EndpointURL))
		}))
	}
	return hops
}

func (p *dynamicGatewaySelector) enrichRouteHop(ctx context.Context, hop *proxyruntimev1.EgressHop, resolve func(context.Context) (string, error)) *proxyruntimev1.EgressHop {
	if hop == nil || resolve == nil {
		return hop
	}
	enrichCtx, cancel := context.WithTimeout(ctx, routeHopEnrichmentTimeout)
	defer cancel()
	results := make(chan *proxyruntimev1.EgressHop, 1)
	go func() {
		enriched := proto.Clone(hop).(*proxyruntimev1.EgressHop)
		ip, err := resolve(enrichCtx)
		if err != nil {
			p.logger.Warn("resolve egress route hop public ip failed", "hop_id", hop.GetHopId(), "error", err)
			results <- enriched
			return
		}
		p.fillRouteHopGeo(enrichCtx, enriched, ip)
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

func (p *dynamicGatewaySelector) fillRouteHopGeo(ctx context.Context, hop *proxyruntimev1.EgressHop, ip string) {
	if hop == nil {
		return
	}
	ip = strings.TrimSpace(ip)
	if net.ParseIP(ip) == nil {
		return
	}
	setRouteHopLabel(hop, "observed_ip", ip)
	geo, err := p.lookupIPGeo(ctx, ip)
	if err != nil {
		p.logger.Warn("resolve egress route hop geo failed", "hop_id", hop.GetHopId(), "observed_ip", ip, "error", err)
		return
	}
	setRouteHopLabel(hop, "country_code", geo.CountryCode)
	setRouteHopLabel(hop, "region", geo.Region)
	setRouteHopLabel(hop, "city", geo.City)
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
