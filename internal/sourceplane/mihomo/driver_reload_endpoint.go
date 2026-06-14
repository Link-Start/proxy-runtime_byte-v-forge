package mihomo

import (
	"context"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func waitForReloadEndpoint(ctx context.Context, endpoint sourceplane.Endpoint) error {
	if strings.TrimSpace(endpoint.Addr) == "" {
		return nil
	}
	return waitForEndpoint(ctx, endpoint.Addr, 3*time.Second)
}
