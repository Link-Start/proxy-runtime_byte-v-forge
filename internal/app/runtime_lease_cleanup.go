package app

import (
	"context"
	"errors"
	"fmt"
)

func (c leaseCoordinator) cleanupPendingLeaseFacts(ctx context.Context) error {
	if c.deps.store == nil {
		return nil
	}
	leases, err := c.deps.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		c.warn("list proxy lease cleanup facts failed", "error", err)
		return err
	}
	cleanupErrors := make([]error, 0)
	for _, lease := range leases {
		if err := ctx.Err(); err != nil {
			cleanupErrors = append(cleanupErrors, err)
			break
		}
		attemptCtx, cancel := context.WithTimeout(ctx, leaseCleanupAttemptTimeout)
		err := c.cleanupPendingLeaseFact(attemptCtx, lease)
		cancel()
		if err != nil {
			c.warn("cleanup proxy lease fact failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
			cleanupErrors = append(cleanupErrors, fmt.Errorf("cleanup lease fact %q: %w", lease.GetLeaseId(), err))
		}
	}
	return errors.Join(cleanupErrors...)
}
