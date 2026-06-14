package app

import (
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func sourcePlaneProxyUserRoutesWithConfigured(settings *runtimeSettingsFile, configured []config.ProxyUserRoute) []dataplane.ProxyUserRoute {
	out := sourcePlaneProxyUserRoutes(settings)
	seen := proxyRouteUsernames(out)
	for _, user := range configured {
		out = appendProxyUserRoute(out, seen, dataplane.ProxyUserRoute{
			ID:        user.ID,
			Username:  user.Username,
			Password:  user.Password,
			Route:     user.Route,
			ProfileID: user.ProfileID,
		})
	}
	return out
}
