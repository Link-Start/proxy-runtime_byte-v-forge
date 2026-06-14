package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type CleanupPendingLeaseInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	LocalProtocol                     string
	Lease                             *proxyruntimev1.ProxyDynamicLease
	IsNotFound                        StoreNotFoundFunc
	ReleaseProvider                   ProviderSessionReleaseAction
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func CleanupPendingLease(ctx context.Context, input CleanupPendingLeaseInput) error {
	return RunCurrentLeaseAction(ctx, CurrentLeaseActionInput{
		Store:      input.Store,
		Locks:      input.Locks,
		Lease:      input.Lease,
		IsNotFound: input.IsNotFound,
		Action: func(ctx context.Context, current *proxyruntimev1.ProxyDynamicLease) error {
			return cleanupCurrentPendingLease(ctx, input, current)
		},
	})
}

func cleanupCurrentPendingLease(ctx context.Context, input CleanupPendingLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !CleanupPending(lease) {
		return nil
	}
	if RouteCleanupPending(lease) {
		if err := cleanupPendingLeaseRoute(ctx, input, lease); err != nil {
			return err
		}
		ClearCleanupPending(lease, true, false)
	}
	if ProviderCleanupPending(lease) {
		if err := cleanupPendingProviderSession(ctx, input, lease); err != nil {
			return err
		}
		ClearCleanupPending(lease, false, true)
	}
	return saveCleanupProgressState(ctx, input, lease)
}

func cleanupPendingLeaseRoute(ctx context.Context, input CleanupPendingLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	return CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return SaveCleanupRetry(ctx, input.Store, lease, "lease route cleanup failed")
		},
	})
}

func cleanupPendingProviderSession(ctx context.Context, input CleanupPendingLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	releaseProvider := func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
		if input.ReleaseProvider == nil {
			return nil
		}
		if releaseErr := input.ReleaseProvider(ctx, lease); releaseErr != nil {
			_ = SaveCleanupRetry(ctx, input.Store, lease, "provider session cleanup failed")
			return releaseErr
		}
		return nil
	}
	return ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, lease, releaseProvider)
}

func saveCleanupProgressState(ctx context.Context, input CleanupPendingLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	err := SaveCleanupProgress(ctx, input.Store, input.Limiter, lease)
	if errors.Is(err, ErrFinalLeaseConcurrencyRelease) {
		observeLease(ctx, input.ObserveFinalConcurrencyReleaseErr, lease)
		return nil
	}
	return err
}
