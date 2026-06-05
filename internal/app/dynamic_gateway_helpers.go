package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func gatewaysForDynamicGateway(settings *runtimeSettingsFile, plan *proxyruntimev1.EgressRoutePlan, providerID string) []accountproxy.Gateway {
	gateways := dynamicIPGateways(settings, providerID)
	gatewayID := strings.TrimSpace(plan.GetDynamicGateway().GetGatewayId())
	if gatewayID == "" {
		return gateways
	}
	for _, gateway := range gateways {
		if firstNonEmpty(gateway.ID, gatewayIDFromEndpointURL(gateway.EndpointURL)) == gatewayID {
			return []accountproxy.Gateway{gateway}
		}
	}
	return gateways
}
