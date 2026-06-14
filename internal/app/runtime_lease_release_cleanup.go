package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) retireLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.RetireLeaseRoute(ctx, leaseapp.RetireLeaseRouteInput{
		Store:         c.deps.store,
		Limiter:       c.deps.providerConcurrency,
		Locks:         c.deps.locks,
		DataPlane:     c.deps.dataPlane,
		LocalProtocol: c.deps.cfg.LocalProtocol,
		Lease:         lease,
		ReleaseProvider: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return c.releaseLeaseProviderSession(ctx, lease)
		},
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
		ObserveFinalConcurrencyReleaseErr: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			_ = ctx
			c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
		},
	})
}
