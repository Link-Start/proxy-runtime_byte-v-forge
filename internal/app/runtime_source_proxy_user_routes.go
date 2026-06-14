package app

import (
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func proxyRouteUsernames(routes []dataplane.ProxyUserRoute) map[string]struct{} {
	seen := map[string]struct{}{}
	for _, route := range routes {
		if username := strings.TrimSpace(route.Username); username != "" {
			seen[username] = struct{}{}
		}
	}
	return seen
}

func appendProxyUserRoute(out []dataplane.ProxyUserRoute, seen map[string]struct{}, route dataplane.ProxyUserRoute) []dataplane.ProxyUserRoute {
	username := strings.TrimSpace(route.Username)
	if username == "" {
		return out
	}
	if _, exists := seen[username]; exists {
		return out
	}
	route.Username = username
	seen[username] = struct{}{}
	return append(out, route)
}
