package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) warn(message string, args ...any) {
	if c.deps.logger != nil {
		c.deps.logger.Warn(message, args...)
	}
}

func (c leaseCoordinator) warnFinalConcurrencyReleaseFailed(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
	_ = ctx
	c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
}

func (c leaseCoordinator) now() time.Time {
	if c.deps.clock != nil {
		return c.deps.clock.Now()
	}
	return time.Now()
}

func (c leaseCoordinator) clearExitCheckCache() {
	if c.deps.exitCheckCache != nil {
		c.deps.exitCheckCache.clear()
	}
}

func (c leaseCoordinator) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if c.deps.closeInUserConnections != nil {
		c.deps.closeInUserConnections(ctx, usernames)
	}
}

func (c leaseCoordinator) providerSessionGatewaysResolver(lease *proxyruntimev1.ProxyDynamicLease) func(context.Context, string) ([]accountproxy.Gateway, error) {
	return func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
		settings, err := c.deps.settings.load(ctx)
		if err != nil {
			return nil, err
		}
		return c.providerSessionGatewaysResolverForSettings(settings, lease)(ctx, providerID)
	}
}

func (c leaseCoordinator) providerSessionGatewaysResolverForSettings(settings *runtimeSettingsFile, lease *proxyruntimev1.ProxyDynamicLease) func(context.Context, string) ([]accountproxy.Gateway, error) {
	return func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
		_ = ctx
		return endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
	}
}

func (c leaseCoordinator) routeLineBindingResolver(settings *runtimeSettingsFile) func(context.Context, string) (string, map[string]string, error) {
	return func(ctx context.Context, accountID string) (string, map[string]string, error) {
		return c.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
	}
}

func (c leaseCoordinator) leaseRouteRetirer() leaseapp.LeaseRouteRetirer {
	return leaseapp.LeaseRouteRetirer{
		Store:                   c.deps.store,
		Limiter:                 c.deps.providerConcurrency,
		Locks:                   c.deps.locks,
		DataPlane:               c.deps.dataPlane,
		Factory:                 c.deps.sessionProviders,
		LocalProtocol:           c.deps.cfg.LocalProtocol,
		ResolveGatewaysForLease: c.providerSessionGatewaysResolver,
		AfterRouteCleanup: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			c.clearExitCheckCache()
			if lease.GetAccountId() == playgroundProfileID {
				c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
			}
		},
		ObserveProviderReleaseFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			_ = ctx
			c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		},
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	}
}

func (c leaseCoordinator) leaseRouteRestorer(settings *runtimeSettingsFile) leaseapp.LeaseRouteRestorer {
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
			return c.providerSessionGatewaysResolverForSettings(settings, lease)
		},
		ResolveLineBinding: c.routeLineBindingResolver(settings),
	}
}

func (c leaseCoordinator) acquiredRouteApplier(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.AcquiredRouteApplier {
	return leaseapp.AcquiredRouteApplier{
		Store:            c.deps.store,
		DataPlane:        c.deps.dataPlane,
		Clock:            c.deps.clock,
		LocalProtocol:    c.deps.cfg.LocalProtocol,
		Managed:          true,
		FallbackProtocol: "http",
		ResolveListener: func(ctx context.Context, accountID string, leaseID string) (leaseapp.Listener, error) {
			return c.deps.leaseListener(ctx, settings, accountID, leaseID)
		},
		ResolveEgress: func(ctx context.Context, listener leaseapp.Listener) (*proxyruntimev1.ProxyEndpoint, error) {
			_ = ctx
			return c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(advertisedHost, listener))
		},
		AfterApply: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) {
			c.clearExitCheckCache()
			if req.GetAccountId() == playgroundProfileID {
				c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
			}
		},
	}
}
