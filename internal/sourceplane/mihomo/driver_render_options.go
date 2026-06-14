package mihomo

import "github.com/byte-v-forge/proxy-runtime/internal/sourceplane"

func (d *Driver) renderConfigProjectionLocked(cfg sourceplane.Config, endpoint sourceplane.Endpoint, dir string) (renderOptions, renderedMihomoConfig, error) {
	options, err := d.configRenderOptionsLocked(cfg, endpoint, dir)
	if err != nil {
		return renderOptions{}, renderedMihomoConfig{}, err
	}
	config, err := renderConfigProjection(options)
	if err != nil {
		return renderOptions{}, renderedMihomoConfig{}, err
	}
	return options, config, nil
}

func (d *Driver) configRenderOptionsLocked(cfg sourceplane.Config, endpoint sourceplane.Endpoint, dir string) (renderOptions, error) {
	nativeConfig, err := loadNativeConfig(dir)
	if err != nil {
		return renderOptions{}, err
	}
	return newRenderOptions(renderOptionsInput{
		Config:          cfg,
		Endpoint:        endpoint,
		ConfigDir:       dir,
		NativeConfig:    nativeConfig,
		BaseConfig:      d.baseCfg,
		SessionRoutes:   d.sessionRoutesLocked(),
		DashboardDir:    firstNonEmpty(d.cfg.DashboardDir, d.baseCfg.DashboardDir),
		DashboardURL:    firstNonEmpty(d.cfg.DashboardURL, d.baseCfg.DashboardURL),
		ControllerAddr:  d.cfg.APIAddr,
		ControllerToken: d.cfg.ControllerSecret,
	}), nil
}
