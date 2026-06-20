package dynamic

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
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

func EndpointsForDynamicIPSelection(settings *proxyruntimev1.ProxyRuntimePersistentSettings, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan, providerID string) []accountproxy.Gateway {
	selected := plan.GetSelectedEndpoint()
	endpoints := Endpoints(settings, selected.GetDynamicProviderId(), providerID)
	endpointID := strings.TrimSpace(plan.GetSelectedEndpoint().GetEndpointId())
	if endpointID == "" {
		return endpoints
	}
	for _, endpoint := range endpoints {
		if appcore.FirstNonEmpty(endpoint.ID, EndpointIDFromURL(endpoint.EndpointURL)) == endpointID {
			return []accountproxy.Gateway{endpoint}
		}
	}
	return endpoints
}
