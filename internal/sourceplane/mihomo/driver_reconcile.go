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
	nativeConfig, err := loadNativeConfig(dir)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	baseOptions := renderOptions{
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
	configFile, err := renderConfig(baseOptions)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	baseSig := signature(data)
	if err := os.MkdirAll(filepath.Join(dir, "providers"), 0o700); err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	configPath := filepath.Join(dir, "config.json")

	restartRequired := !d.running || d.lastEndpoint != endpoint
	baseChanged := d.baseSig != baseSig
	baseReloaded := false
	if restartRequired {
		if err := writeConfigData(configPath, data); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		d.stopLocked()
		if err := d.startLocked(ctx, dir, configPath); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		if err := waitForEndpoint(ctx, endpoint.Addr, 3*time.Second); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		baseReloaded = true
	} else if baseChanged {
		if err := d.reloadConfigDataLocked(ctx, configPath, data, endpoint); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
		baseReloaded = true
	}

	finalOptions := baseOptions
	finalConfig, err := renderConfig(finalOptions)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	finalData, err := json.MarshalIndent(finalConfig, "", "  ")
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	finalSig := signature(finalData)
	d.desiredSig = finalSig
	if baseReloaded || finalSig != d.signature {
		if err := d.reloadConfigDataLocked(ctx, configPath, finalData, endpoint); err != nil {
			d.lastError = err.Error()
			return nil, err
		}
	}
	d.signature = finalSig
	d.baseSig = baseSig
	d.configPath = configPath
	d.lastEndpoint = endpoint
	d.lastError = ""
	return nil, nil
}
