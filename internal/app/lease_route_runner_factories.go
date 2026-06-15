package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type leaseRouteRetirerFactory struct {
	deps             leaseCoordinatorDependencies
	settings         leaseapp.SettingsAdapter[*runtimeSettingsFile]
	sideEffects      leaseRouteSideEffects
	observeFinalSlot leaseapp.LeaseObserver
}

func (f leaseRouteRetirerFactory) New() leaseapp.LeaseRouteRetirer {
	return leaseapp.LeaseRouteRetirer{
		Store:                             f.deps.store,
		Limiter:                           f.deps.providerConcurrency,
		Locks:                             f.deps.locks,
		DataPlane:                         f.deps.dataPlane,
		Factory:                           f.deps.sessionProviders,
		LocalProtocol:                     f.deps.cfg.LocalProtocol,
		ResolveGatewaysForLease:           f.settings.ProviderGatewaysResolver,
		AfterRouteCleanup:                 f.afterRouteCleanup,
		ObserveProviderReleaseFailure:     f.observeProviderReleaseFailure,
		ObserveFinalConcurrencyReleaseErr: f.observeFinalSlot,
	}
}

func (f leaseRouteRetirerFactory) afterRouteCleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
	f.sideEffects.afterRouteChange(ctx, lease.GetAccountId())
}

func (f leaseRouteRetirerFactory) observeProviderReleaseFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
	_ = ctx
	warnLeaseProviderSessionReleaseFailed(f.deps.logger, lease)
}

type leaseReleaseRunnerFactory struct {
	deps   leaseCoordinatorDependencies
	retire leaseapp.LeaseRouteRetirer
}

func (f leaseReleaseRunnerFactory) New() leaseapp.ReleaseRunner {
	return leaseapp.ReleaseRunner{
		Store:      f.deps.store,
		Locks:      f.deps.locks,
		IsNotFound: isStoreNotFound,
		Retire:     f.retire.Retire,
	}
}

type leaseRouteRestorerFactory struct {
	deps     leaseCoordinatorDependencies
	settings *runtimeSettingsFile
	adapter  leaseapp.SettingsAdapter[*runtimeSettingsFile]
}

func (f leaseRouteRestorerFactory) New() leaseapp.LeaseRouteRestorer {
	return leaseapp.LeaseRouteRestorer{
		Limiter:            f.deps.providerConcurrency,
		Store:              f.deps.store,
		DataPlane:          f.deps.dataPlane,
		Factory:            f.deps.sessionProviders,
		DefaultTTL:         leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:          providerAccountConcurrencyTTLBuffer,
		SlotReleaseTimeout: leaseRestoreSlotReleaseTimeout,
		LocalProtocol:      f.deps.cfg.LocalProtocol,
		Limit:              f.limit,
		ResolveGateways:    f.resolveGateways,
		ResolveLineBinding: f.adapter.RouteLineBindingResolver(f.settings),
	}
}

func (f leaseRouteRestorerFactory) limit(lease *proxyruntimev1.ProxyDynamicLease) uint32 {
	return dynamicProviderConcurrencyLimit(f.settings, leaseapp.DynamicProviderID(lease), leaseapp.ConcurrencyPolicy(lease))
}

func (f leaseRouteRestorerFactory) resolveGateways(lease *proxyruntimev1.ProxyDynamicLease) leaseapp.ProviderSessionGatewaysResolver {
	return f.adapter.ProviderGatewaysResolverForSettings(f.settings, lease)
}

func warnLeaseProviderSessionReleaseFailed(logger leaseapp.Logger, lease *proxyruntimev1.ProxyDynamicLease) {
	if logger == nil {
		return
	}
	logger.Warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
}
