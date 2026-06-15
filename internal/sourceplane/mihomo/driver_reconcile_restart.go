package mihomo

import (
	"context"
	"path/filepath"
)

func (d *Driver) restartConfigProjectionLocked(ctx context.Context, input configProjectionApplyInput) error {
	if err := writeConfigData(input.configPath, input.config.data); err != nil {
		return configProjectionStageError("write restart config", err)
	}
	d.stopLocked()
	if err := d.startLocked(ctx, filepath.Dir(input.configPath), input.configPath); err != nil {
		return configProjectionStageError("start process", err)
	}
	if err := waitForReloadEndpoint(ctx, input.endpoint); err != nil {
		return configProjectionStageError("wait listener", err)
	}
	return nil
}
