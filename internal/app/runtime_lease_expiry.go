package app

import (
	"context"
	"errors"
	"fmt"
)

func (c leaseCoordinator) expireDueLeaseFacts(ctx context.Context) error {
	if c.deps.store == nil {
		return nil
	}
	leases, err := c.deps.store.ExpiredActiveLeaseFacts(ctx)
	if err != nil {
		return err
	}
	expireErrors := make([]error, 0)
	for _, lease := range leases {
		if err := ctx.Err(); err != nil {
			expireErrors = append(expireErrors, err)
			break
		}
		attemptCtx, cancel := context.WithTimeout(ctx, leaseCleanupAttemptTimeout)
		err := c.expireLeaseFact(attemptCtx, lease)
		cancel()
		if err != nil {
			c.warn("expire proxy lease failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
			expireErrors = append(expireErrors, fmt.Errorf("expire lease fact %q: %w", lease.GetLeaseId(), err))
		}
	}
	return errors.Join(expireErrors...)
}
