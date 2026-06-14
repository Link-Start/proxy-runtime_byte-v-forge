package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
