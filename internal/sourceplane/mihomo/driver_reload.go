package mihomo

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) reloadConfigDataLocked(ctx context.Context, canonicalPath string, data []byte, endpoint sourceplane.Endpoint) error {
	candidatePath := filepath.Join(filepath.Dir(canonicalPath), "config.candidate.json")
	if err := writeConfigData(candidatePath, data); err != nil {
		return err
	}
	if err := d.reloadLocked(ctx, candidatePath); err != nil {
		return err
	}
	if strings.TrimSpace(endpoint.Addr) != "" {
		if err := waitForEndpoint(ctx, endpoint.Addr, 3*time.Second); err != nil {
			return err
		}
	}
	return writeConfigData(canonicalPath, data)
}
