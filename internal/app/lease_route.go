package app

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/byte-v-forge/proxy-gateway/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
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

func (f leaseRouteRetirerFactory) afterRouteCleanup(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) {
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
		IsNotFound: store.IsNotFound,
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
		DefaultTTL:         kernel.DefaultDynamicIPStickyTTL,
		TTLBuffer:          providerAccountConcurrencyTTLBuffer,
		SlotReleaseTimeout: leaseRestoreSlotReleaseTimeout,
		LocalProtocol:      f.deps.cfg.LocalProtocol,
		Limit:              f.limit,
		ResolveGateways:    f.resolveGateways,
		ResolveLineBinding: f.adapter.RouteLineBindingResolver(f.settings),
	}
}

func (f leaseRouteRestorerFactory) limit(lease *proxygatewayv1.ProxyDynamicLease) uint32 {
	return dynamicProviderConcurrencyLimit(f.settings, leaseapp.DynamicProviderID(lease), leaseapp.ConcurrencyPolicy(lease))
}

func (f leaseRouteRestorerFactory) resolveGateways(lease *proxygatewayv1.ProxyDynamicLease) leaseapp.ProviderSessionGatewaysResolver {
	return f.adapter.ProviderGatewaysResolverForSettings(f.settings, lease)
}

func warnLeaseProviderSessionReleaseFailed(logger leaseapp.Logger) leaseapp.LeaseErrorObserver {
	return func(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease, err error) {
		_ = ctx
		if logger == nil || lease == nil {
			return
		}
		logger.Warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId(), "error_type", appcore.ErrorLogType(err))
	}
}

type leaseRouteSideEffects struct {
	exitCheckCache         *proxycheck.ExitCheckCache
	closeInUserConnections leaseConnectionCleanupFunc
}

func (e leaseRouteSideEffects) afterRouteChange(ctx context.Context, accountID string) {
	e.clearExitCheckCache()
	if accountID == kernel.PlaygroundProfileID {
		e.closePlaygroundConnections(ctx)
	}
}

func (e leaseRouteSideEffects) clearExitCheckCache() {
	if e.exitCheckCache != nil {
		e.exitCheckCache.Clear()
	}
}

func (e leaseRouteSideEffects) closePlaygroundConnections(ctx context.Context) {
	if e.closeInUserConnections != nil {
		e.closeInUserConnections(ctx, []string{kernel.PlaygroundUsername})
	}
}
