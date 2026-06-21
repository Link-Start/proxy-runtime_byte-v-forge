package ten24

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func (p *Provider) Descriptor() *proxygatewayv1.ProxyProviderDescriptor {
	capabilities := []proxygatewayv1.ProxyCapability{
		proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
	}
	upstreamKinds := []proxygatewayv1.ProxyUpstreamKind{}
	rotationModes := []proxygatewayv1.ProxyRotationMode{}
	if p.cfg.APIURL != "" {
		capabilities = append(capabilities,
			proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_API_POOL,
			proxygatewayv1.ProxyCapability_PROXY_CAPABILITY_POOL_REFRESH,
		)
		upstreamKinds = append(upstreamKinds, proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_PROXY_POOL)
		rotationModes = append(rotationModes,
			proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST,
			proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_SCHEDULED_POOL_REFRESH,
		)
	}
	return &proxygatewayv1.ProxyProviderDescriptor{
		ProviderId:    p.Name(),
		DisplayName:   "1024Proxy API pool",
		Capabilities:  capabilities,
		Protocols:     []proxygatewayv1.ProxyProtocol{protocolEnum(defaultProtocol(p.cfg.Protocol))},
		UpstreamKinds: upstreamKinds,
		RotationModes: rotationModes,
	}
}
