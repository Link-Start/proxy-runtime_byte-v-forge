package lease

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type workerOperation func(context.Context) error

func (a *Application) RestoreActive(ctx context.Context) error {
	return a.runWorkerOperation(ctx, "restore_active", func(ctx context.Context) error {
		return a.worker.RestoreActive(ctx)
	})
}

func (a *Application) ExpireDue(ctx context.Context) error {
	return a.runWorkerOperation(ctx, "expire_due", func(ctx context.Context) error {
		return a.worker.ExpireDue(ctx)
	})
}

func (a *Application) CleanupPending(ctx context.Context) error {
	return a.runWorkerOperation(ctx, "cleanup_pending", func(ctx context.Context) error {
		return a.worker.CleanupPending(ctx)
	})
}

func (a *Application) Cleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return a.runWorkerOperation(ctx, "cleanup_one", func(ctx context.Context) error {
		return a.worker.Cleanup(ctx, lease)
	}, leaseWorkerLogFields(lease)...)
}

func (a *Application) runWorkerOperation(ctx context.Context, operation string, run workerOperation, fields ...any) error {
	if a == nil || a.worker == nil {
		return fmt.Errorf("lease worker is required")
	}
	startedAt := a.now()
	err := run(ctx)
	logFields := append([]any{"operation", operation, "duration_ms", a.sinceMilliseconds(startedAt)}, fields...)
	if err != nil {
		a.warn("lease worker operation failed", append(logFields, "error_type", errorType(err))...)
		return err
	}
	a.info("lease worker operation finished", logFields...)
	return nil
}

func leaseWorkerLogFields(lease *proxyruntimev1.ProxyDynamicLease) []any {
	if lease == nil {
		return nil
	}
	return []any{
		"lease_id", lease.GetLeaseId(),
		"account_id", lease.GetAccountId(),
		"purpose", lease.GetPurpose(),
		"provider_account_key", lease.GetProviderAccountId(),
	}
}
