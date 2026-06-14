package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
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
		return err
	}
	return writeConfigData(canonicalPath, data)
}
