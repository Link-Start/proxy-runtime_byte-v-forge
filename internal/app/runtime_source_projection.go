package app

import (
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type sourcePlaneConfigInput struct {
	Settings          *runtimeSettingsFile
	Pool              []provider.Node
	LocalAddr         string
	HealthCheckURL    string
	HealthCheckPeriod time.Duration
	HealthCheckWait   time.Duration
	DashboardDir      string
	DashboardURL      string
	ProxyUsers        []config.ProxyUserRoute
}

func sourcePlaneDataPlaneConfig(input sourcePlaneConfigInput) dataplane.Config {
	return dataplane.Config{
		EgressProfiles:    sourcePlaneEgressProfiles(input.Settings),
		Endpoint:          sourceplane.Endpoint{Addr: input.LocalAddr, Protocol: "socks5"},
		HealthCheckURL:    input.HealthCheckURL,
		HealthCheckPeriod: input.HealthCheckPeriod,
		HealthCheckWait:   input.HealthCheckWait,
		DashboardDir:      input.DashboardDir,
		DashboardURL:      input.DashboardURL,
		Pool:              input.Pool,
		ProxyUsers:        sourcePlaneProxyUserRoutesWithConfigured(input.Settings, input.ProxyUsers),
	}
}

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
