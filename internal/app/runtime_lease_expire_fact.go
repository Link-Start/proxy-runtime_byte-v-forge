package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.RunCurrentLeaseAction(ctx, leaseapp.CurrentLeaseActionInput{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		Lease:      lease,
		IsNotFound: isStoreNotFound,
		Action: func(ctx context.Context, current *proxyruntimev1.ProxyDynamicLease) error {
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
		},
	})
}
