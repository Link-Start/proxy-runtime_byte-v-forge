package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ExpireLeaseInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	LocalProtocol                     string
	Lease                             *proxyruntimev1.ProxyDynamicLease
	IsNotFound                        StoreNotFoundFunc
	Now                               time.Time
	ReleaseProvider                   ProviderSessionReleaseAction
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func ExpireLease(ctx context.Context, input ExpireLeaseInput) error {
	return RunCurrentLeaseAction(ctx, CurrentLeaseActionInput{
		Store:      input.Store,
		Locks:      input.Locks,
		Lease:      input.Lease,
		IsNotFound: input.IsNotFound,
		Action: func(ctx context.Context, current *proxyruntimev1.ProxyDynamicLease) error {
			return expireCurrentLease(ctx, input, current)
		},
	})
}

func expireCurrentLease(ctx context.Context, input ExpireLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !NeedsExpiryCleanup(lease, input.Now) {
		return nil
	}
	if err := expireLeaseRoute(ctx, input, lease); err != nil {
		return err
	}
	if err := expireLeaseProviderSession(ctx, input, lease); err != nil {
		return err
	}
	return saveExpiredFinalState(ctx, input, lease)
}

func expireLeaseRoute(ctx context.Context, input ExpireLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	return CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return SaveExpiredCleanupFailure(ctx, input.Store, lease, true, false, "expired lease route cleanup failed")
		},
	})
}

func expireLeaseProviderSession(ctx context.Context, input ExpireLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	releaseProvider := func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
		if input.ReleaseProvider == nil {
			return nil
		}
		if releaseErr := input.ReleaseProvider(ctx, lease); releaseErr != nil {
			_ = SaveExpiredCleanupFailure(ctx, input.Store, lease, false, true, "expired provider session cleanup failed")
			return releaseErr
		}
		return nil
	}
	if err := ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, lease, releaseProvider); err != nil {
		_ = SaveExpiredCleanupFailure(ctx, input.Store, lease, false, true, "expired provider session cleanup lock failed")
		return err
	}
	return nil
}

func saveExpiredFinalState(ctx context.Context, input ExpireLeaseInput, lease *proxyruntimev1.ProxyDynamicLease) error {
	err := SaveExpiredFinalLeaseState(ctx, input.Store, input.Limiter, lease)
	if errors.Is(err, ErrFinalLeaseConcurrencyRelease) {
		observeLease(ctx, input.ObserveFinalConcurrencyReleaseErr, lease)
		return nil
	}
	return err
}
