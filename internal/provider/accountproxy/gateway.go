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

func GatewayProtocol(_ Gateway, fallback string) string {
	return defaultProtocol(fallback, "socks5")
}
