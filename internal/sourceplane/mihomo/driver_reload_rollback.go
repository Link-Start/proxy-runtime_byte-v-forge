package mihomo

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) rollbackConfigReloadLocked(ctx context.Context, canonicalPath string, endpoint sourceplane.Endpoint, cause error) error {
	if cause == nil {
		return nil
	}
	if rollbackErr := d.reloadLocked(ctx, canonicalPath); rollbackErr != nil {
		return fmt.Errorf("mihomo config reload failed: %w; rollback reload failed: %v", cause, rollbackErr)
	}
	if rollbackErr := waitForReloadEndpoint(ctx, endpoint); rollbackErr != nil {
		return fmt.Errorf("mihomo config reload failed: %w; rollback endpoint check failed: %v", cause, rollbackErr)
	}
	return cause
}
