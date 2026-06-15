package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type WorkerProcessor struct {
	Store              OrchestrationStore
	Clock              Clock
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
