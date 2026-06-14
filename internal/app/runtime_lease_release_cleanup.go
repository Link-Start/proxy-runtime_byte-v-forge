package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) retireLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	if err := c.cleanupLeaseRoute(ctx, lease, func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
		return c.saveLeaseReleaseCleanupFailure(ctx, lease, true, false, "lease route cleanup failed")
	}); err != nil {
		return err
	}
	c.clearExitCheckCache()
	if lease.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	releaseErr := leaseapp.ReleaseLeaseProviderSessionWithLock(ctx, c.deps.locks, lease, c.releaseLeaseProviderSession)
	if releaseErr != nil {
		c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		if err := c.saveLeaseReleaseCleanupFailure(ctx, lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return c.saveLeaseReleased(ctx, lease)
}
