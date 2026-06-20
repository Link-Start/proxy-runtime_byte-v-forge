package lease

import (
	"context"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
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

type WorkerBatchListErrorObserver func(error)

type WorkerBatchInput struct {
	Store       OrchestrationStore
	Timeout     time.Duration
	ShouldRun   BatchPredicate
	Process     BatchProcessor
	Observe     BatchErrorObserver
	ObserveList WorkerBatchListErrorObserver
}

func ProcessCleanupPendingFacts(ctx context.Context, input WorkerBatchInput) error {
	return processWorkerBatch(ctx, input, func(ctx context.Context, store OrchestrationStore) ([]*proxyruntimev1.ProxyDynamicLease, error) {
		return store.CleanupPendingLeaseFacts(ctx)
	}, "cleanup lease fact")
}

func ProcessExpiredActiveFacts(ctx context.Context, input WorkerBatchInput) error {
	return processWorkerBatch(ctx, input, func(ctx context.Context, store OrchestrationStore) ([]*proxyruntimev1.ProxyDynamicLease, error) {
		return store.ExpiredActiveLeaseFacts(ctx)
	}, "expire lease fact")
}

func ProcessRestorableActiveFacts(ctx context.Context, input WorkerBatchInput, now time.Time) error {
	input.ShouldRun = activeLeasePredicate(now)
	return processWorkerBatch(ctx, input, func(ctx context.Context, store OrchestrationStore) ([]*proxyruntimev1.ProxyDynamicLease, error) {
		return store.ListRestorableLeaseFacts(ctx)
	}, "restore lease route")
}

func processWorkerBatch(ctx context.Context, input WorkerBatchInput, list func(context.Context, OrchestrationStore) ([]*proxyruntimev1.ProxyDynamicLease, error), errorPrefix string) error {
	if input.Store == nil || list == nil {
		return nil
	}
	leases, err := list(ctx, input.Store)
	if err != nil {
		if input.ObserveList != nil {
			input.ObserveList(err)
		}
		return err
	}
	return ProcessLeaseBatch(ctx, BatchInput{
		Leases:      leases,
		Timeout:     input.Timeout,
		ErrorPrefix: errorPrefix,
		ShouldRun:   input.ShouldRun,
		Process:     input.Process,
		Observe:     input.Observe,
	})
}

func activeLeasePredicate(now time.Time) BatchPredicate {
	return func(lease *proxyruntimev1.ProxyDynamicLease) bool {
		return ActiveAt(lease, now)
	}
}

type WorkerProcessor struct {
	Store              OrchestrationStore
	Clock              clock.Clock
	RestoreTimeout     time.Duration
	CleanupTimeout     time.Duration
	Restore            BatchProcessor
	Expire             BatchProcessor
	CleanupPendingOne  BatchProcessor
	ObserveRestore     BatchErrorObserver
	ObserveRestoreList WorkerBatchListErrorObserver
	ObserveExpire      BatchErrorObserver
	ObserveCleanup     BatchErrorObserver
	ObserveCleanupList WorkerBatchListErrorObserver
}

func (w WorkerProcessor) RestoreActive(ctx context.Context) error {
	return ProcessRestorableActiveFacts(ctx, WorkerBatchInput{
		Store:       w.Store,
		Timeout:     w.RestoreTimeout,
		Process:     w.Restore,
		Observe:     w.ObserveRestore,
		ObserveList: w.ObserveRestoreList,
	}, w.now().UTC())
}

func (w WorkerProcessor) ExpireDue(ctx context.Context) error {
	return ProcessExpiredActiveFacts(ctx, WorkerBatchInput{
		Store:   w.Store,
		Timeout: w.CleanupTimeout,
		Process: w.Expire,
		Observe: w.ObserveExpire,
	})
}

func (w WorkerProcessor) CleanupPending(ctx context.Context) error {
	return ProcessCleanupPendingFacts(ctx, WorkerBatchInput{
		Store:       w.Store,
		Timeout:     w.CleanupTimeout,
		Process:     w.CleanupPendingOne,
		Observe:     w.ObserveCleanup,
		ObserveList: w.ObserveCleanupList,
	})
}

func (w WorkerProcessor) Cleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if w.CleanupPendingOne == nil {
		return nil
	}
	return w.CleanupPendingOne(ctx, lease)
}

func (w WorkerProcessor) now() time.Time {
	if w.Clock != nil {
		return w.Clock.Now()
	}
	return time.Now()
}
