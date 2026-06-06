package app

import (
	"context"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (r *Runtime) dataPlaneConfig(ctx context.Context) (dataplane.Config, error) {
	settings, err := r.settings.load(ctx)
	if err != nil {
		return dataplane.Config{}, err
	}
	pool, err := r.dynamicProfilePool(ctx, settings)
	if err != nil {
		return dataplane.Config{}, err
	}
	r.setDynamicProfilePoolSnapshot(pool)
	return dataplane.Config{
		EgressProfiles:    sourcePlaneEgressProfiles(settings),
		Endpoint:          sourceplane.Endpoint{Addr: r.cfg.LocalAddr, Protocol: "socks5"},
		HealthCheckURL:    r.cfg.Mihomo.HealthCheckURL,
		HealthCheckPeriod: r.cfg.Mihomo.HealthCheckInterval,
		HealthCheckWait:   r.cfg.Mihomo.HealthCheckTimeout,
		DashboardDir:      r.cfg.Mihomo.DashboardDir,
		DashboardURL:      r.cfg.Mihomo.DashboardURL,
		Pool:              pool,
		ProxyUsers:        r.proxyUserRoutes(settings),
	}, nil
}

func (r *Runtime) proxyUserRoutes(settings *runtimeSettingsFile) []dataplane.ProxyUserRoute {
	out := sourcePlaneProxyUserRoutes(settings)
	seen := proxyRouteUsernames(out)
	for _, user := range r.cfg.ProxyUsers {
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
