package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func protocolEnum(value string) proxyruntimev1.ProxyProtocol {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http", "https":
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP
	case "socks5", "socks5h":
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	default:
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
}

func endpointsForDynamicIPSelection(settings *runtimeSettingsFile, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan, providerID string) []accountproxy.Gateway {
	selected := plan.GetSelectedEndpoint()
	endpoints := dynamicIPEndpoints(settings, selected.GetDynamicProviderId(), providerID)
	endpointID := strings.TrimSpace(plan.GetSelectedEndpoint().GetEndpointId())
	if endpointID == "" {
		return endpoints
	}
	for _, endpoint := range endpoints {
		if firstNonEmpty(endpoint.ID, endpointIDFromURL(endpoint.EndpointURL)) == endpointID {
			return []accountproxy.Gateway{endpoint}
		}
	}
	return endpoints
}
