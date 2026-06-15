package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) reconcileLocked(ctx context.Context, cfg sourceplane.Config) ([]provider.Node, error) {
	endpoint, err := normalizeEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("normalize endpoint", err))
	}
	dir, err := d.ensureConfigDir()
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("prepare config directory", err))
	}
	baseOptions, baseConfig, err := d.renderConfigProjectionLocked(cfg, endpoint, dir)
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("render base config", err))
	}
	if err := ensureProviderConfigDir(dir); err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("prepare provider config directory", err))
	}
	configPath := runtimeConfigPath(dir)

	baseApply, err := d.applyBaseConfigProjectionLocked(ctx, configPath, baseConfig, endpoint)
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("apply base config", err))
	}

	finalOptions := baseOptions
	finalConfig, err := renderConfigProjection(finalOptions)
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("render final config", err))
	}
	if err := d.applyFinalConfigProjectionLocked(ctx, configPath, finalConfig, endpoint, baseApply); err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("apply final config", err))
	}
	d.recordAppliedConfigProjection(configPath, endpoint, baseConfig, finalConfig)
	return nil, nil
}
