package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
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
