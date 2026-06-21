package app

import (
	"context"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/config"
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

func (r *Runtime) dataPlaneConfig(ctx context.Context) (dataplane.Config, error) {
	settings, err := r.settings.Load(ctx)
	if err != nil {
		return dataplane.Config{}, err
	}
	pool, err := r.dynamicProfilePool(ctx, settings)
	if err != nil {
		return dataplane.Config{}, err
	}
	r.setDynamicProfilePoolSnapshot(pool)
	return sourcePlaneDataPlaneConfig(sourcePlaneConfigInput{
		Settings:          settings,
		Pool:              pool,
		LocalAddr:         r.cfg.LocalAddr,
		HealthCheckURL:    r.cfg.Mihomo.HealthCheckURL,
		HealthCheckPeriod: r.cfg.Mihomo.HealthCheckInterval,
		HealthCheckWait:   r.cfg.Mihomo.HealthCheckTimeout,
		DashboardDir:      r.cfg.Mihomo.DashboardDir,
		DashboardURL:      r.cfg.Mihomo.DashboardURL,
		ProxyUsers:        r.cfg.ProxyUsers,
	}), nil
}

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
