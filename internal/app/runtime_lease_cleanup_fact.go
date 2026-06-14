package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !leaseapp.HasLeaseID(lease) {
		return nil
	}
	return leaseapp.RunAccountAction(ctx, c.deps.locks, lease.GetAccountId(), func(ctx context.Context) error {
		current, err := c.deps.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil {
			if isStoreNotFound(err) {
				return nil
			}
			return err
		}
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
	})
}
