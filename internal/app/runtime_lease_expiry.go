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
			if err := r.leaseCoordinator.cleanupPendingLeaseFacts(ctx); err != nil {
				r.logger.Warn("cleanup pending proxy leases failed", "error", err)
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
	return r.leaseLocks.WithAccountLock(ctx, lease.GetAccountId(), func(ctx context.Context) error {
		current, err := r.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil {
			if isStoreNotFound(err) {
				return nil
			}
			return err
		}
		if leaseapp.ActiveAt(current, time.Now().UTC()) {
			return nil
		}
		if current.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
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
		if strings.TrimSpace(current.GetProviderAccountId()) != "" {
			if err := r.leaseLocks.WithProviderAccountLock(ctx, current.GetProviderAccountId(), releaseProvider); err != nil {
				_ = c.saveLeaseExpiredCleanupFailure(ctx, current, false, true, "expired provider session cleanup lock failed")
				return err
			}
		} else if err := releaseProvider(ctx); err != nil {
			return err
		}
		return c.saveLeaseExpired(ctx, current)
	})
}
