package ten24

import (
	"net/url"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func setQuery(query url.Values, key string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		query.Set(key, value)
	}
}

func defaultProtocol(protocol string) string {
	switch strings.TrimSpace(protocol) {
	case "socks5":
		return "socks5"
	default:
		return "http"
	}
}

func protocolEnum(protocol string) proxyruntimev1.ProxyProtocol {
	switch defaultProtocol(protocol) {
	case "socks5":
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	default:
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP
	}
}
