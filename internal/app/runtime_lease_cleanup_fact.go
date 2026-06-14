package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil || strings.TrimSpace(lease.GetLeaseId()) == "" {
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
		if !leaseapp.CleanupPending(current) {
			return nil
		}
		if leaseapp.RouteCleanupPending(current) {
			if err := c.deleteLeaseRoute(ctx, current); err != nil {
				_ = c.saveLeaseCleanupRetry(ctx, current, "lease route cleanup failed")
				return err
			}
			leaseapp.ClearCleanupPending(current, true, false)
		}
		if leaseapp.ProviderCleanupPending(current) {
			releaseProvider := func(ctx context.Context) error {
				if releaseErr := c.releaseLeaseProviderSession(ctx, current); releaseErr != nil {
					_ = c.saveLeaseCleanupRetry(ctx, current, "provider session cleanup failed")
					return releaseErr
				}
				return nil
			}
			if err := leaseapp.WithProviderAccountLock(ctx, c.deps.locks, current.GetProviderAccountId(), releaseProvider); err != nil {
				return err
			}
			leaseapp.ClearCleanupPending(current, false, true)
		}
		if !leaseapp.CleanupPending(current) {
			switch leaseapp.CleanupFinalStatus(current) {
			case leaseapp.CleanupFinalExpired:
				return c.saveLeaseExpired(ctx, current)
			case leaseapp.CleanupFinalReleased:
				return c.saveLeaseReleased(ctx, current)
			}
			return c.deps.store.SaveLeaseFact(ctx, current)
		}
		return c.deps.store.SaveLeaseFact(ctx, current)
	})
}
