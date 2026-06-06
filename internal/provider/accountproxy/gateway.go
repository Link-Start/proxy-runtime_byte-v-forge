package accountproxy

import (
	"net/url"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func gatewayForPolicy(definition Definition, _ *proxyruntimev1.ProxySessionPolicy) (Gateway, bool) {
	return defaultGateway(definition)
}

func gatewayEndpointHost(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return value
}

func GatewayProtocol(gateway Gateway, fallback string) string {
	endpointURL := strings.TrimSpace(gateway.EndpointURL)
	if parsed, err := url.Parse(endpointURL); err == nil && parsed.Scheme != "" {
		return defaultProtocol(parsed.Scheme, fallback)
	}
	return defaultProtocol(fallback, "socks5")
}
