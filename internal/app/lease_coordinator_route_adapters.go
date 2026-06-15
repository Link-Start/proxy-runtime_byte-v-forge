package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) leaseRouteRetirer() leaseapp.LeaseRouteRetirer {
	settings := c.settingsAdapter()
	sideEffects := c.routeSideEffects()
	return leaseapp.LeaseRouteRetirer{
		Store:                   c.deps.store,
		Limiter:                 c.deps.providerConcurrency,
		Locks:                   c.deps.locks,
		DataPlane:               c.deps.dataPlane,
		Factory:                 c.deps.sessionProviders,
		LocalProtocol:           c.deps.cfg.LocalProtocol,
		ResolveGatewaysForLease: settings.ProviderGatewaysResolver,
		AfterRouteCleanup: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			sideEffects.afterRouteChange(ctx, lease.GetAccountId())
		},
		ObserveProviderReleaseFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			_ = ctx
			c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		},
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	}
}

func (c leaseCoordinator) releaseRunner() leaseapp.ReleaseRunner {
	retirer := c.leaseRouteRetirer()
	return leaseapp.ReleaseRunner{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		IsNotFound: isStoreNotFound,
		Retire:     retirer.Retire,
	}
}

func (c leaseCoordinator) leaseRouteRestorer(settings *runtimeSettingsFile) leaseapp.LeaseRouteRestorer {
	settingsAdapter := c.settingsAdapter()
	return leaseapp.LeaseRouteRestorer{
		Limiter:            c.deps.providerConcurrency,
		Store:              c.deps.store,
		DataPlane:          c.deps.dataPlane,
		Factory:            c.deps.sessionProviders,
		DefaultTTL:         leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:          providerAccountConcurrencyTTLBuffer,
		SlotReleaseTimeout: leaseRestoreSlotReleaseTimeout,
		LocalProtocol:      c.deps.cfg.LocalProtocol,
		Limit: func(lease *proxyruntimev1.ProxyDynamicLease) uint32 {
			return dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), leaseapp.ConcurrencyPolicy(lease))
		},
		ResolveGateways: func(lease *proxyruntimev1.ProxyDynamicLease) leaseapp.ProviderSessionGatewaysResolver {
			return settingsAdapter.ProviderGatewaysResolverForSettings(settings, lease)
		},
		ResolveLineBinding: settingsAdapter.RouteLineBindingResolver(settings),
	}
}

func (c leaseCoordinator) acquiredRouteApplier(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.AcquiredRouteApplier {
	sideEffects := c.routeSideEffects()
	endpoint := leaseAcquiredEndpointAdapter{
		settings:              settings,
		advertisedHost:        advertisedHost,
		leaseListener:         c.deps.leaseListener,
		localListenerEndpoint: c.deps.localListenerEndpoint,
		sessionAdvertisedHost: c.deps.sessionAdvertisedHost,
	}
	return leaseapp.AcquiredRouteApplier{
		Store:            c.deps.store,
		DataPlane:        c.deps.dataPlane,
		Clock:            c.deps.clock,
		LocalProtocol:    c.deps.cfg.LocalProtocol,
		Managed:          true,
		FallbackProtocol: "http",
		ResolveListener:  endpoint.ResolveListener,
		ResolveEgress:    endpoint.ResolveEgress,
		AfterApply: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) {
			sideEffects.afterRouteChange(ctx, req.GetAccountId())
		},
	}
}
