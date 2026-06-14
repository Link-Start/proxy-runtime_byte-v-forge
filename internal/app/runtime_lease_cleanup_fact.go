package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return leaseapp.RunCurrentLeaseAction(ctx, leaseapp.CurrentLeaseActionInput{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		Lease:      lease,
		IsNotFound: isStoreNotFound,
		Action: func(ctx context.Context, current *proxyruntimev1.ProxyDynamicLease) error {
			if !leaseapp.CleanupPending(current) {
				return nil
			}
			if leaseapp.RouteCleanupPending(current) {
				if err := c.cleanupLeaseRoute(ctx, current, func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
					return c.saveLeaseCleanupRetry(ctx, lease, "lease route cleanup failed")
				}); err != nil {
					return err
				}
				leaseapp.ClearCleanupPending(current, true, false)
			}
			if leaseapp.ProviderCleanupPending(current) {
				releaseProvider := func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
					if releaseErr := c.releaseLeaseProviderSession(ctx, lease); releaseErr != nil {
						_ = c.saveLeaseCleanupRetry(ctx, lease, "provider session cleanup failed")
						return releaseErr
					}
					return nil
				}
				if err := leaseapp.ReleaseLeaseProviderSessionWithLock(ctx, c.deps.locks, current, releaseProvider); err != nil {
					return err
				}
				leaseapp.ClearCleanupPending(current, false, true)
			}
			return c.saveLeaseCleanupProgress(ctx, current)
		},
	})
}
