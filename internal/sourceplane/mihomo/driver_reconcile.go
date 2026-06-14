package mihomo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type renderedMihomoConfig struct {
	data      []byte
	signature string
}

func (d *Driver) reconcileLocked(ctx context.Context, cfg sourceplane.Config) ([]provider.Node, error) {
	endpoint, err := normalizeEndpoint(cfg.Endpoint)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	dir, err := d.ensureConfigDir()
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	baseOptions, baseConfig, err := d.renderConfigProjectionLocked(cfg, endpoint, dir)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "providers"), 0o700); err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	configPath := filepath.Join(dir, "config.json")

	baseReloaded, err := d.applyBaseConfigProjectionLocked(ctx, configPath, baseConfig, endpoint)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}

	finalOptions := baseOptions
	finalConfig, err := renderConfigProjection(finalOptions)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	d.desiredSig = finalConfig.signature
	if baseReloaded || finalConfig.signature != d.signature {
		if err := d.reloadConfigDataLocked(ctx, configPath, finalConfig.data, endpoint); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
	}
	d.signature = finalConfig.signature
	d.baseSig = baseConfig.signature
	d.configPath = configPath
	d.lastEndpoint = endpoint
	d.lastError = ""
	return nil, nil
}

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

func renderConfigProjection(options renderOptions) (renderedMihomoConfig, error) {
	configFile, err := renderConfig(options)
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	return renderedMihomoConfig{data: data, signature: signature(data)}, nil
}

func (d *Driver) applyBaseConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint) (bool, error) {
	restartRequired := !d.running || d.lastEndpoint != endpoint
	baseChanged := d.baseSig != config.signature
	if restartRequired {
		if err := writeConfigData(configPath, config.data); err != nil {
			return false, err
		}
		d.stopLocked()
		if err := d.startLocked(ctx, filepath.Dir(configPath), configPath); err != nil {
			return false, err
		}
		if err := waitForEndpoint(ctx, endpoint.Addr, 3*time.Second); err != nil {
			return false, err
		}
		return true, nil
	}
	if baseChanged {
		if err := d.reloadConfigDataLocked(ctx, configPath, config.data, endpoint); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
