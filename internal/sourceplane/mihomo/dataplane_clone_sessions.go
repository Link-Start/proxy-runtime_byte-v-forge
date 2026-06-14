package mihomo

import "github.com/byte-v-forge/proxy-runtime/internal/dataplane"

func cloneSessionRoute(route dataplane.SessionRoute) dataplane.SessionRoute {
	route.Pool = cloneProviderNodes(route.Pool)
	return route
}
