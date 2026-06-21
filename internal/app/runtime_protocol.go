package app

import (
	"fmt"
	"net/url"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func protocolFromName(protocol string) proxygatewayv1.ProxyProtocol {
	if protocol == "socks5" {
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_HTTP
}

func protocolName(protocol proxygatewayv1.ProxyProtocol) string {
	if protocol == proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5 {
		return "socks5"
	}
	return "http"
}

func protocolFromURL(proxyURL *url.URL) proxygatewayv1.ProxyProtocol {
	if proxyURL != nil && proxyURL.Scheme == "socks5" {
		return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	}
	return proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_HTTP
}

func portFromURL(proxyURL *url.URL) uint32 {
	if proxyURL == nil || proxyURL.Port() == "" {
		return 0
	}
	port, _ := parsePort(proxyURL.Port())
	return port
}

func parsePort(portValue string) (uint32, error) {
	var port uint32
	_, err := fmt.Sscanf(portValue, "%d", &port)
	return port, err
}
