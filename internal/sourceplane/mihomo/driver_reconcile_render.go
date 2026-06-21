package mihomo

import "github.com/byte-v-forge/proxy-gateway/internal/sourceplane"

type baseConfigProjection struct {
	endpoint      sourceplane.Endpoint
	configDir     string
	configPath    string
	renderOptions renderOptions
	config        renderedMihomoConfig
}

func (d *Driver) buildBaseConfigProjectionLocked(cfg sourceplane.Config) (baseConfigProjection, error) {
	endpoint, err := normalizeEndpoint(cfg.Endpoint)
	if err != nil {
		return baseConfigProjection{}, configProjectionStageError("normalize endpoint", err)
	}
	configDir, err := d.ensureConfigDir()
	if err != nil {
		return baseConfigProjection{}, configProjectionStageError("prepare config directory", err)
	}
	options, config, err := d.renderConfigProjectionLocked(cfg, endpoint, configDir)
	if err != nil {
		return baseConfigProjection{}, configProjectionStageError("render base config", err)
	}
	return baseConfigProjection{
		endpoint:      endpoint,
		configDir:     configDir,
		configPath:    runtimeConfigPath(configDir),
		renderOptions: options,
		config:        config,
	}, nil
}

func prepareProviderConfigProjectionDir(configDir string) error {
	if err := ensureProviderConfigDir(configDir); err != nil {
		return configProjectionStageError("prepare provider config directory", err)
	}
	return nil
}

func buildFinalConfigProjection(options renderOptions) (renderedMihomoConfig, error) {
	config, err := renderConfigProjection(options)
	if err != nil {
		return renderedMihomoConfig{}, configProjectionStageError("render final config", err)
	}
	return config, nil
}

func (p baseConfigProjection) applyInput(config renderedMihomoConfig) configProjectionApplyInput {
	return configProjectionApplyInput{
		configPath: p.configPath,
		config:     config,
		endpoint:   p.endpoint,
	}
}
