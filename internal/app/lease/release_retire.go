package lease

import (
	"context"
	"errors"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

var ErrReleaseRetireActionRequired = errors.New("release retire action is required")

type ReleaseRetireAction func(context.Context, *proxygatewayv1.ProxyDynamicLease) error

type ReleaseRetireInput struct {
	Store      OrchestrationStore
	Locks      LockManager
	Lease      *proxygatewayv1.ProxyDynamicLease
	IsNotFound StoreNotFoundFunc
	Retire     ReleaseRetireAction
}

func RetireReleaseLease(ctx context.Context, input ReleaseRetireInput) (*proxygatewayv1.ProxyDynamicLease, error) {
	lease := input.Lease
	if !ReleaseNeedsRouteRetire(lease) {
		return lease, nil
	}
	if input.Retire == nil {
		return lease, ErrReleaseRetireActionRequired
	}
	err := WithAccountLock(ctx, input.Locks, lease.GetAccountId(), func(ctx context.Context) error {
		current, err := RefreshReleaseLease(ctx, input.Store, lease, input.IsNotFound)
		lease = current
		if err != nil || !ReleaseNeedsRouteRetire(lease) {
			return err
		}
		return input.Retire(ctx, lease)
	})
	return lease, err
}

type LeaseObserver func(context.Context, *proxygatewayv1.ProxyDynamicLease)

type LeaseErrorObserver func(context.Context, *proxygatewayv1.ProxyDynamicLease, error)

type ProviderSessionGatewaysResolverFactory func(*proxygatewayv1.ProxyDynamicLease) ProviderSessionGatewaysResolver

type LeaseRouteRetirer struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	LocalProtocol                     string
	ResolveGatewaysForLease           ProviderSessionGatewaysResolverFactory
	AfterRouteCleanup                 LeaseObserver
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

type RetireLeaseRouteInput struct {
	Store                             OrchestrationStore
	Limiter                           ProviderAccountConcurrencyLimiter
	Locks                             LockManager
	DataPlane                         DataPlaneApplier
	Factory                           SessionProviderFactory
	LocalProtocol                     string
	Lease                             *proxygatewayv1.ProxyDynamicLease
	ResolveGateways                   ProviderSessionGatewaysResolver
	AfterRouteCleanup                 LeaseObserver
	ObserveProviderReleaseFailure     LeaseErrorObserver
	ObserveFinalConcurrencyReleaseErr LeaseObserver
}

func (r LeaseRouteRetirer) Retire(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	return RetireLeaseRoute(ctx, RetireLeaseRouteInput{
		Store:                             r.Store,
		Limiter:                           r.Limiter,
		Locks:                             r.Locks,
		DataPlane:                         r.DataPlane,
		Factory:                           r.Factory,
		LocalProtocol:                     r.LocalProtocol,
		Lease:                             lease,
		ResolveGateways:                   r.resolveGateways(lease),
		AfterRouteCleanup:                 r.AfterRouteCleanup,
		ObserveProviderReleaseFailure:     r.ObserveProviderReleaseFailure,
		ObserveFinalConcurrencyReleaseErr: r.ObserveFinalConcurrencyReleaseErr,
	})
}

func (r LeaseRouteRetirer) resolveGateways(lease *proxygatewayv1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	if r.ResolveGatewaysForLease == nil {
		return nil
	}
	return r.ResolveGatewaysForLease(lease)
}

func RetireLeaseRoute(ctx context.Context, input RetireLeaseRouteInput) error {
	if input.Lease == nil {
		return nil
	}
	if err := CleanupLeaseRoute(ctx, RouteCleanupInput{
		DataPlane:     input.DataPlane,
		Lease:         input.Lease,
		LocalProtocol: input.LocalProtocol,
		RecordFailure: func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
			return SaveReleaseCleanupFailure(ctx, input.Store, lease, true, false, "lease route cleanup failed")
		},
	}); err != nil {
		return err
	}
	observeLease(ctx, input.AfterRouteCleanup, input.Lease)
	releaseErr := ReleaseLeaseProviderSessionWithLock(ctx, input.Locks, ProviderSessionReleaseInput{
		Store:           input.Store,
		Factory:         input.Factory,
		Lease:           input.Lease,
		ResolveGateways: input.ResolveGateways,
	})
	if releaseErr != nil {
		observeLeaseErr(ctx, input.ObserveProviderReleaseFailure, input.Lease, releaseErr)
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

func observeLease(ctx context.Context, observe LeaseObserver, lease *proxygatewayv1.ProxyDynamicLease) {
	if observe != nil {
		observe(ctx, lease)
	}
}

func observeLeaseErr(ctx context.Context, observe LeaseErrorObserver, lease *proxygatewayv1.ProxyDynamicLease, err error) {
	if observe != nil {
		observe(ctx, lease, err)
	}
}
