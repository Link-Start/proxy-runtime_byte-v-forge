package dynamic

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func protocolEnum(value string) proxygatewayv1.ProxyProtocol {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http", "https":
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_HTTP
	case "socks5", "socks5h":
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	default:
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
}

func EndpointsForDynamicIPSelection(settings *proxygatewayv1.ProxyGatewayPersistentSettings, plan *proxygatewayv1.ProxyDynamicIPSelectionPlan, providerID string) []accountproxy.Gateway {
	selected := plan.GetSelectedEndpoint()
	endpoints := Endpoints(settings, selected.GetDynamicProviderId(), providerID)
	endpointID := strings.TrimSpace(plan.GetSelectedEndpoint().GetEndpointId())
	if endpointID == "" {
		return endpoints
	}
	for _, endpoint := range endpoints {
		if appcore.FirstNonEmpty(endpoint.ID, kernel.EndpointIDFromURL(endpoint.EndpointURL)) == endpointID {
			return []accountproxy.Gateway{endpoint}
		}
	}
	return endpoints
}
