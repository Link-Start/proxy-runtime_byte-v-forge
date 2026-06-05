package ten24

import proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func (p *Provider) Descriptor() *proxyruntimev1.ProxyProviderDescriptor {
	capabilities := []proxyruntimev1.ProxyCapability{
		proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
	}
	upstreamKinds := []proxyruntimev1.ProxyUpstreamKind{}
	rotationModes := []proxyruntimev1.ProxyRotationMode{}
	if p.cfg.APIURL != "" {
		capabilities = append(capabilities,
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_API_POOL,
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_POOL_REFRESH,
		)
		upstreamKinds = append(upstreamKinds, proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_PROXY_POOL)
		rotationModes = append(rotationModes,
			proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST,
			proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_SCHEDULED_POOL_REFRESH,
		)
	}
	return &proxyruntimev1.ProxyProviderDescriptor{
		ProviderId:    p.Name(),
		DisplayName:   "1024Proxy API pool",
		Capabilities:  capabilities,
		Protocols:     []proxyruntimev1.ProxyProtocol{protocolEnum(defaultProtocol(p.cfg.Protocol))},
		UpstreamKinds: upstreamKinds,
		RotationModes: rotationModes,
	}
}

func (p *Provider) Sources() []*proxyruntimev1.ProxySourceDescriptor {
	if p.cfg.APIURL == "" {
		return nil
	}
	return []*proxyruntimev1.ProxySourceDescriptor{{
		SourceId:    "1024proxy-api-pool",
		ProviderId:  p.Name(),
		DisplayName: "1024Proxy API pool",
		Kind:        proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_API_POOL,
		Capabilities: []proxyruntimev1.ProxyCapability{
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_API_POOL,
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_POOL_REFRESH,
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
		},
		Protocols: []proxyruntimev1.ProxyProtocol{protocolEnum(defaultProtocol(p.cfg.Protocol))},
		Model: &proxyruntimev1.ProxySourceDescriptor_ApiPool{
			ApiPool: &proxyruntimev1.ProxyAPIPoolSourceDescriptor{},
		},
	}}
}
