package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/jackc/pgx/v5"
)

const leaseExpirySweepInterval = 30 * time.Second

func (r *Runtime) leaseExpiryLoop(ctx context.Context) {
	ticker := time.NewTicker(leaseExpirySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.leaseCoordinator.expireDueLeaseFacts(ctx); err != nil {
				r.logger.Warn("expire proxy leases failed", "error", err)
			}
		}
	}
}

func (c leaseCoordinator) expireDueLeaseFacts(ctx context.Context) error {
	r := c.runtime
	if r.store == nil {
		return nil
	}
	leases, err := r.store.ExpiredActiveLeaseFacts(ctx)
	if err != nil {
		return err
	}
	expireErrors := make([]error, 0)
	for _, lease := range leases {
		if err := c.expireLeaseFact(ctx, lease); err != nil {
			r.logger.Warn("expire proxy lease failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
			expireErrors = append(expireErrors, fmt.Errorf("expire lease fact %q: %w", lease.GetLeaseId(), err))
		}
	}
	return errors.Join(expireErrors...)
}

func (c leaseCoordinator) expireLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	if lease == nil || strings.TrimSpace(lease.GetLeaseId()) == "" {
		return nil
	}
	lock, err := r.leaseLocks.LockAccount(ctx, lease.GetAccountId())
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	current, err := r.store.LeaseFactByID(ctx, lease.GetLeaseId())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if leaseActive(current, time.Now().UTC()) {
		return nil
	}
	if current.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
		return nil
	}
	if err := c.deleteLeaseRoute(ctx, current); err != nil {
		_ = c.saveLeaseExpiredCleanupFailure(ctx, current, true, false, "expired lease route cleanup failed")
		return err
	}
	var providerLock leaseRuntimeLock
	if strings.TrimSpace(current.GetProviderAccountId()) != "" {
		providerLock, err = r.leaseLocks.LockProviderAccount(ctx, current.GetProviderAccountId())
		if err != nil {
			_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup lock failed")
			return err
		}
	}
	releaseErr := c.releaseLeaseProviderSession(ctx, current)
	if providerLock != nil {
		_ = providerLock.Unlock(ctx)
	}
	if releaseErr != nil {
		_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup failed")
		return releaseErr
	}
	return c.saveLeaseExpired(ctx, current)
}
