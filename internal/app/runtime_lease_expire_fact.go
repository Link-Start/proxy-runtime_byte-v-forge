package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
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
		if leaseapp.ActiveAt(current, c.now().UTC()) {
			return nil
		}
		if !leaseapp.HasActiveStatus(current) {
			return nil
		}
		if err := c.deleteLeaseRoute(ctx, current); err != nil {
			_ = c.saveLeaseExpiredCleanupFailure(ctx, current, true, false, "expired lease route cleanup failed")
			return err
		}
		releaseProvider := func(ctx context.Context) error {
			if releaseErr := c.releaseLeaseProviderSession(ctx, current); releaseErr != nil {
				_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup failed")
				return releaseErr
			}
			return nil
		}
		if err := leaseapp.WithProviderAccountLock(ctx, c.deps.locks, current.GetProviderAccountId(), releaseProvider); err != nil {
			_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup lock failed")
			return err
		}
		return c.saveLeaseExpired(ctx, current)
	})
}
