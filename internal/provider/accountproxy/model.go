package accountproxy

import (
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func descriptor(definition Definition, gateways []Gateway) *proxygatewayv1.ProxyProviderDescriptor {
	definition.Gateways = gateways
	return &proxygatewayv1.ProxyProviderDescriptor{
		ProviderId:    definition.ProviderID,
		DisplayName:   definition.DisplayName,
		Capabilities:  capabilities(definition),
		Protocols:     protocols(definition),
		MinStickyTtl:  stickyDuration(minStickyMinutes),
		MaxStickyTtl:  stickyDuration(maxStickyMinutes),
		UpstreamKinds: upstreamKinds(definition),
		RotationModes: rotationModes(definition),
	}
}

func capabilities(definition Definition) []proxygatewayv1.ProxyCapability {
	out := []proxygatewayv1.ProxyCapability{}
	if len(definition.Gateways) == 0 {
		return out
	}
	out = append(out, proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_STICKY_SESSION, proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY, proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_DYNAMIC_LEASE)
	if definition.UsernameParameterSession {
		out = append(out, proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_ACTIVE_SESSION_ROTATION, proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_USERNAME_PARAMETER_SESSION)
	}
	return out
}

func upstreamKinds(definition Definition) []proxygatewayv1.ProxyUpstreamKind {
	if len(definition.Gateways) == 0 {
		return nil
	}
	return []proxygatewayv1.ProxyUpstreamKind{proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP}
}

func rotationModes(definition Definition) []proxygatewayv1.ProxyRotationMode {
	if len(definition.Gateways) == 0 {
		return nil
	}
	out := []proxygatewayv1.ProxyRotationMode{proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION}
	if definition.UsernameParameterSession {
		out = append(out, proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST)
	}
	return out
}

func protocols(definition Definition) []proxygatewayv1.ProxyProtocol {
	values := definition.Protocols
	if len(values) == 0 {
		values = []string{definition.DefaultProtocol}
	}
	out := make([]proxygatewayv1.ProxyProtocol, 0, len(values))
	seen := map[proxygatewayv1.ProxyProtocol]struct{}{}
	for _, value := range values {
		protocol := protocolEnumWithDefault(value, definition.DefaultProtocol)
		if protocol == proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED {
			continue
		}
		if _, exists := seen[protocol]; exists {
			continue
		}
		seen[protocol] = struct{}{}
		out = append(out, protocol)
	}
	if len(out) == 0 {
		out = append(out, protocolEnumWithDefault(definition.DefaultProtocol, "socks5"))
	}
	return out
}

func defaultGateway(definition Definition) (Gateway, bool) {
	for _, gateway := range definition.Gateways {
		if strings.TrimSpace(gateway.EndpointURL) != "" {
			return gateway, true
		}
	}
	return Gateway{}, false
}

func stickyDuration(minutes int) *durationpb.Duration {
	return durationpb.New(time.Duration(clampStickyMinutes(minutes)) * time.Minute)
}
