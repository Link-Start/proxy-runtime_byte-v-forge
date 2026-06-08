package accountproxy

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func descriptor(definition Definition, gateways []Gateway) *proxyruntimev1.ProxyProviderDescriptor {
	definition.Gateways = gateways
	return &proxyruntimev1.ProxyProviderDescriptor{
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

func capabilities(definition Definition) []proxyruntimev1.ProxyCapability {
	out := []proxyruntimev1.ProxyCapability{}
	if len(definition.Gateways) == 0 {
		return out
	}
	out = append(out, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_STICKY_SESSION, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_DYNAMIC_LEASE)
	if definition.UsernameParameterSession {
		out = append(out, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_ACTIVE_SESSION_ROTATION, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_USERNAME_PARAMETER_SESSION)
	}
	return out
}

func upstreamKinds(definition Definition) []proxyruntimev1.ProxyUpstreamKind {
	if len(definition.Gateways) == 0 {
		return nil
	}
	return []proxyruntimev1.ProxyUpstreamKind{proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP}
}

func rotationModes(definition Definition) []proxyruntimev1.ProxyRotationMode {
	if len(definition.Gateways) == 0 {
		return nil
	}
	out := []proxyruntimev1.ProxyRotationMode{proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION}
	if definition.UsernameParameterSession {
		out = append(out, proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST)
	}
	return out
}

func protocols(definition Definition) []proxyruntimev1.ProxyProtocol {
	values := definition.Protocols
	if len(values) == 0 {
		values = []string{definition.DefaultProtocol}
	}
	out := make([]proxyruntimev1.ProxyProtocol, 0, len(values))
	seen := map[proxyruntimev1.ProxyProtocol]struct{}{}
	for _, value := range values {
		protocol := protocolEnumWithDefault(value, definition.DefaultProtocol)
		if protocol == proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED {
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
