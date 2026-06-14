package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !leaseapp.HasLeaseID(lease) {
		return nil
	}
	return leaseapp.WithAccountLock(ctx, c.deps.locks, lease.GetAccountId(), func(ctx context.Context) error {
		current, err := c.deps.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil {
			if isStoreNotFound(err) {
				return nil
			}
			return err
		}
		if !leaseapp.NeedsExpiryCleanup(current, c.now().UTC()) {
			return nil
		}
		if err := c.cleanupLeaseRoute(ctx, current, func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			return c.saveLeaseExpiredCleanupFailure(ctx, lease, true, false, "expired lease route cleanup failed")
		}); err != nil {
			return err
		}
		releaseProvider := func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
			if releaseErr := c.releaseLeaseProviderSession(ctx, lease); releaseErr != nil {
				_ = c.saveLeaseExpiredCleanupFailure(ctx, lease, false, true, "expired provider session cleanup failed")
				return releaseErr
			}
			return nil
		}
		if err := leaseapp.ReleaseLeaseProviderSessionWithLock(ctx, c.deps.locks, current, releaseProvider); err != nil {
			_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup lock failed")
			return err
		}
		return c.saveLeaseExpired(ctx, current)
	})
}
