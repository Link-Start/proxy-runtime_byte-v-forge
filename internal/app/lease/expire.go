package lease

import (
	"context"
	"errors"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
)

type ExpireLeaseRunner struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	Clock                             clock.Clock
	LocalProtocol                     string
	IsNotFound                        StoreNotFoundFunc
	ResolveGatewaysForLease           ProviderSessionGatewaysResolverFactory
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func (r ExpireLeaseRunner) Expire(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	return ExpireLease(ctx, ExpireLeaseInput{
		Store:                             r.Store,
		Limiter:                           r.Limiter,
		Locks:                             r.Locks,
		DataPlane:                         r.DataPlane,
		Factory:                           r.Factory,
		LocalProtocol:                     r.LocalProtocol,
		Lease:                             lease,
		IsNotFound:                        r.IsNotFound,
		Now:                               r.now().UTC(),
		ResolveGateways:                   r.resolveGateways(lease),
		ObserveProviderReleaseFailure:     r.ObserveProviderReleaseFailure,
		ObserveFinalConcurrencyReleaseErr: r.ObserveFinalConcurrencyReleaseErr,
	})
}

func (r ExpireLeaseRunner) resolveGateways(lease *proxygatewayv1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGatewaysForLease == nil {
		return nil
	}
	return r.ResolveGatewaysForLease(lease)
}

func (r ExpireLeaseRunner) now() time.Time {
	if r.Clock != nil {
		return r.Clock.Now()
	}
	return time.Now()
}

type ExpireLeaseInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	LocalProtocol                     string
	Lease                             *proxygatewayv1.ProxyDynamicLease
	IsNotFound                        StoreNotFoundFunc
	Now                               time.Time
	ResolveGateways                   ProviderSessionGatewaysResolver
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func ExpireLease(ctx context.Context, input ExpireLeaseInput) error {
	return RunCurrentLeaseAction(ctx, CurrentLeaseActionInput{
		Store:      input.Store,
		Locks:      input.Locks,
		Lease:      input.Lease,
		IsNotFound: input.IsNotFound,
		Action: func(ctx context.Context, current *proxygatewayv1.ProxyDynamicLease) error {
			return expireCurrentLease(ctx, input, current)
		},
	})
}

func expireCurrentLease(ctx context.Context, input ExpireLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
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

func expireLeaseRoute(ctx context.Context, input ExpireLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	return CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
			return SaveExpiredCleanupFailure(ctx, input.Store, lease, true, false, "expired lease route cleanup failed")
		},
	})
}

func expireLeaseProviderSession(ctx context.Context, input ExpireLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	err := ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, ProviderSessionReleaseInput{
		Store:           input.Store,
		Factory:         input.Factory,
		Lease:           lease,
		ResolveGateways: input.ResolveGateways,
		RecordFailure: func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease, err error) error {
			observeLeaseErr(ctx, input.ObserveProviderReleaseFailure, lease, err)
			return SaveExpiredCleanupFailure(ctx, input.Store, lease, false, true, "expired provider session cleanup failed")
		},
	})
	if err != nil {
		_ = SaveExpiredCleanupFailure(ctx, input.Store, lease, false, true, "expired provider session cleanup lock failed")
		return err
	}
	return nil
}

func saveExpiredFinalState(ctx context.Context, input ExpireLeaseInput, lease *proxygatewayv1.ProxyDynamicLease) error {
	err := SaveExpiredFinalLeaseState(ctx, input.Store, input.Limiter, lease)
	if errors.Is(err, ErrFinalLeaseConcurrencyRelease) {
		observeLease(ctx, input.ObserveFinalConcurrencyReleaseErr, lease)
		return nil
	}
	return err
}
