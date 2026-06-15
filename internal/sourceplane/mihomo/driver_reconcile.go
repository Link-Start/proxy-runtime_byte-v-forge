package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) reconcileLocked(ctx context.Context, cfg sourceplane.Config) ([]provider.Node, error) {
	baseProjection, err := d.buildBaseConfigProjectionLocked(cfg)
	if err != nil {
		return nil, d.recordConfigProjectionError(err)
	}
	if err := prepareProviderConfigProjectionDir(baseProjection.configDir); err != nil {
		return nil, d.recordConfigProjectionError(err)
	}

	baseApply, err := d.applyBaseConfigProjectionLocked(ctx, baseProjection.configPath, baseProjection.config, baseProjection.endpoint)
	if err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("apply base config", err))
	}

	finalConfig, err := buildFinalConfigProjection(baseProjection.renderOptions)
	if err != nil {
		return nil, d.recordConfigProjectionError(err)
	}
	if err := d.applyFinalConfigProjectionLocked(ctx, baseProjection.configPath, finalConfig, baseProjection.endpoint, baseApply); err != nil {
		return nil, d.recordConfigProjectionError(configProjectionStageError("apply final config", err))
	}
	d.recordAppliedConfigProjection(baseProjection.configPath, baseProjection.endpoint, baseProjection.config, finalConfig)
	return nil, nil
}
