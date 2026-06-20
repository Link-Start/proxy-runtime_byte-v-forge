package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
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
		ObserveProviderReleaseFailure:     warnLeaseProviderSessionReleaseFailed(f.deps.logger),
		ObserveFinalConcurrencyReleaseErr: f.observeFinalConcurrencyReleaseFailure(),
	}
}

func (f leaseRouteRetirerFactory) afterRouteCleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
	f.sideEffects.afterRouteChange(ctx, lease.GetAccountId())
}

func (f leaseRouteRetirerFactory) observeFinalConcurrencyReleaseFailure() leaseapp.LeaseObserver {
	if f.observeFinalSlot != nil {
		return f.observeFinalSlot
	}
	return warnFinalLeaseConcurrencyReleaseFailed(f.deps.logger)
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

func warnLeaseProviderSessionReleaseFailed(logger leaseapp.Logger) leaseapp.LeaseErrorObserver {
	return func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, err error) {
		_ = ctx
		if logger == nil || lease == nil {
			return
		}
		logger.Warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId(), "error_type", appcore.ErrorLogType(err))
	}
}
