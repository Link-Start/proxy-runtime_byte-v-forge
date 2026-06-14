package mihomo

import (
	"context"

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
	baseOptions, baseConfig, err := d.renderConfigProjectionLocked(cfg, endpoint, dir)
	if err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	if err := ensureProviderConfigDir(dir); err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	configPath := runtimeConfigPath(dir)

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
	if err := d.applyFinalConfigProjectionLocked(ctx, configPath, finalConfig, endpoint, baseReloaded); err != nil {
		d.lastError = err.Error()
		return nil, err
	}
	d.recordAppliedConfigProjection(configPath, endpoint, baseConfig, finalConfig)
	return nil, nil
}
