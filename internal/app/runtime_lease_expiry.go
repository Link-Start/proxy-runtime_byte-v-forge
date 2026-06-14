package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const (
	leaseExpirySweepInterval   = 30 * time.Second
	leaseCleanupAttemptTimeout = 20 * time.Second
)

func (r *Runtime) leaseExpiryLoop(ctx context.Context) {
	ticker := time.NewTicker(leaseExpirySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.service().leases.ExpireDueLeaseFacts(ctx); err != nil {
				r.logger.Warn("expire proxy leases failed", "error", err)
			}
			if err := r.service().leases.CleanupPendingLeaseFacts(ctx); err != nil {
				r.logger.Warn("cleanup pending proxy leases failed", "error", err)
			}
		}
	}
}

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
