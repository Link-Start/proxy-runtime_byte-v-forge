package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

func (d *Driver) reloadConfigDataLocked(ctx context.Context, canonicalPath string, data []byte, endpoint sourceplane.Endpoint) error {
	candidatePath := candidateConfigPath(canonicalPath)
	if err := writeConfigData(candidatePath, data); err != nil {
		return err
	}
	if err := d.reloadLocked(ctx, candidatePath); err != nil {
		return err
	}
	if err := waitForReloadEndpoint(ctx, endpoint); err != nil {
		return d.rollbackConfigReloadLocked(ctx, canonicalPath, endpoint, err)
	}
	if err := writeConfigData(canonicalPath, data); err != nil {
		return d.rollbackConfigReloadLocked(ctx, canonicalPath, endpoint, err)
	}
	return nil
}
