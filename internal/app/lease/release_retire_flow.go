package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type LeaseObserver func(context.Context, *proxyruntimev1.ProxyDynamicLease)

type RetireLeaseRouteInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	LocalProtocol                     string
	Lease                             *proxyruntimev1.ProxyDynamicLease
	ReleaseProvider                   ProviderSessionReleaseAction
	AfterRouteCleanup                 LeaseObserver
	ObserveProviderReleaseFailure     LeaseObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func RetireLeaseRoute(ctx context.Context, input RetireLeaseRouteInput) error {
	if input.Lease == nil {
		return nil
	}
	if err := CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         input.Lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return SaveReleaseCleanupFailure(ctx, input.Store, lease, true, false, "lease route cleanup failed")
		},
	}); err != nil {
		return err
	}
	observeLease(ctx, input.AfterRouteCleanup, input.Lease)
	releaseErr := ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, input.Lease, input.ReleaseProvider)
	if releaseErr != nil {
		observeLease(ctx, input.ObserveProviderReleaseFailure, input.Lease)
		if err := SaveReleaseCleanupFailure(ctx, input.Store, input.Lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return saveReleasedFinalState(ctx, input)
}

func saveReleasedFinalState(ctx context.Context, input RetireLeaseRouteInput) error {
	err := SaveReleasedFinalLeaseState(ctx, input.Store, input.Limiter, input.Lease)
	if errors.Is(err, ErrFinalLeaseConcurrencyRelease) {
		observeLease(ctx, input.ObserveFinalConcurrencyReleaseErr, input.Lease)
		return nil
	}
	return err
}

func observeLease(ctx context.Context, observe LeaseObserver, lease *proxyruntimev1.ProxyDynamicLease) {
	if observe != nil {
		observe(ctx, lease)
	}
}
