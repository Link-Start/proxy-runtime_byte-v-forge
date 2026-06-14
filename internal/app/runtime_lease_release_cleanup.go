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
	if err := leaseapp.DeleteLeaseRoute(ctx, c.deps.dataPlane, lease, c.deps.cfg.LocalProtocol); err != nil {
		_ = c.saveLeaseReleaseCleanupFailure(ctx, lease, true, false, "lease route cleanup failed")
		return err
	}
	c.clearExitCheckCache()
	if lease.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	releaseErr := c.releaseLeaseProviderSessionWithLock(ctx, lease)
	if releaseErr != nil {
		c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		if err := c.saveLeaseReleaseCleanupFailure(ctx, lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return c.saveLeaseReleased(ctx, lease)
}

func (c leaseCoordinator) releaseLeaseProviderSessionWithLock(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.WithProviderAccountLock(ctx, c.deps.locks, lease.GetProviderAccountId(), func(ctx context.Context) error {
		return c.releaseLeaseProviderSession(ctx, lease)
	})
}
