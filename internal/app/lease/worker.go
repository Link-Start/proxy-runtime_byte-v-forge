package lease

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a *Application) RestoreActive(ctx context.Context) error {
	if a == nil || a.worker == nil {
		return fmt.Errorf("lease worker is required")
	}
	return a.worker.RestoreActive(ctx)
}

func (a *Application) ExpireDue(ctx context.Context) error {
	if a == nil || a.worker == nil {
		return fmt.Errorf("lease worker is required")
	}
	return a.worker.ExpireDue(ctx)
}

func (a *Application) CleanupPending(ctx context.Context) error {
	if a == nil || a.worker == nil {
		return fmt.Errorf("lease worker is required")
	}
	return a.worker.CleanupPending(ctx)
}

func (a *Application) Cleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if a == nil || a.worker == nil {
		return fmt.Errorf("lease worker is required")
	}
	return a.worker.Cleanup(ctx, lease)
}
