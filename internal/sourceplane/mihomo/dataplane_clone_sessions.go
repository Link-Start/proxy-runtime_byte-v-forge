package mihomo

import "github.com/byte-v-forge/proxy-gateway/internal/dataplane"

func cloneSessionRoute(route dataplane.SessionRoute) dataplane.SessionRoute {
	route.Pool = cloneProviderNodes(route.Pool)
	return route
}
