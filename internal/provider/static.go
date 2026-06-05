package provider

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/proxyurl"
)

type Static struct {
	nodes []Node
}

func NewStatic(rawProxies []string) (*Static, error) {
	nodes := make([]Node, 0, len(rawProxies))
	for index, raw := range rawProxies {
		proxyURL, err := proxyurl.Parse(raw, "http")
		if err != nil {
			return nil, fmt.Errorf("parse static proxy %d: %w", index, err)
		}
		nodes = append(nodes, Node{
			ID:           fmt.Sprintf("static-%d", index),
			URL:          proxyURL,
			ProviderID:   StaticProviderID,
			UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_SIMPLE_PROXY,
			RotationMode: proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_NONE,
			Labels: map[string]string{
				"mode": "simple_proxy",
			},
		})
	}
	return &Static{nodes: nodes}, nil
}

func (s *Static) Name() string {
	return StaticProviderID
}

func (s *Static) Descriptor() *proxyruntimev1.ProxyProviderDescriptor {
	return &proxyruntimev1.ProxyProviderDescriptor{
		ProviderId:  s.Name(),
		DisplayName: "Static proxies",
		Capabilities: []proxyruntimev1.ProxyCapability{
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
		},
		Protocols: []proxyruntimev1.ProxyProtocol{
			proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP,
			proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5,
		},
		UpstreamKinds: []proxyruntimev1.ProxyUpstreamKind{
			proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_SIMPLE_PROXY,
		},
		RotationModes: []proxyruntimev1.ProxyRotationMode{
			proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_NONE,
		},
	}
}

func (s *Static) Sources() []*proxyruntimev1.ProxySourceDescriptor {
	return []*proxyruntimev1.ProxySourceDescriptor{{
		SourceId:    "static",
		ProviderId:  s.Name(),
		DisplayName: "Static proxies",
		Kind:        proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_FIXED_PROXY,
		Capabilities: []proxyruntimev1.ProxyCapability{
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
		},
		Protocols: []proxyruntimev1.ProxyProtocol{
			proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP,
			proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5,
		},
		Model: &proxyruntimev1.ProxySourceDescriptor_FixedProxy{
			FixedProxy: &proxyruntimev1.ProxyFixedSourceDescriptor{EndpointCount: uint32(len(s.nodes))},
		},
	}}
}

func (s *Static) Fetch(context.Context) ([]Node, error) {
	nodes := make([]Node, len(s.nodes))
	for index, node := range s.nodes {
		clonedURL := *node.URL
		nodes[index] = node
		nodes[index].URL = &clonedURL
		nodes[index].Labels = cloneLabels(node.Labels)
	}
	return nodes, nil
}
