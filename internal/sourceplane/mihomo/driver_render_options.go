package mihomo

import "github.com/byte-v-forge/proxy-runtime/internal/sourceplane"

func (d *Driver) renderConfigProjectionLocked(cfg sourceplane.Config, endpoint sourceplane.Endpoint, dir string) (renderOptions, renderedMihomoConfig, error) {
	nativeConfig, err := loadNativeConfig(dir)
	if err != nil {
		return renderOptions{}, renderedMihomoConfig{}, err
	}
	options := renderOptions{
		EgressProfiles:      cfg.EgressProfiles,
		Endpoint:            endpoint,
		ConfigDir:           dir,
		NativeConfig:        nativeConfig,
		APIAddr:             d.cfg.APIAddr,
		ControllerSecret:    d.cfg.ControllerSecret,
		DashboardDir:        firstNonEmpty(d.cfg.DashboardDir, d.baseCfg.DashboardDir),
		DashboardURL:        firstNonEmpty(d.cfg.DashboardURL, d.baseCfg.DashboardURL),
		HealthCheckURL:      cfg.HealthCheckURL,
		HealthCheckInterval: cfg.HealthCheckInterval,
		HealthCheckTimeout:  cfg.HealthCheckTimeout,
		BasePool:            d.baseCfg.Pool,
		ProxyUsers:          d.baseCfg.ProxyUsers,
		SessionRoutes:       d.sessionRoutesLocked(),
	}
	config, err := renderConfigProjection(options)
	if err != nil {
		return renderOptions{}, renderedMihomoConfig{}, err
	}
	return options, config, nil
}
