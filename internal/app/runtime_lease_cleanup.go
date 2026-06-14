package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const (
	leaseCleanupFinalFailed   = "failed"
	leaseCleanupFinalExpired  = "expired"
	leaseCleanupFinalReleased = "released"
)

func (c leaseCoordinator) cleanupPendingLeaseFacts(ctx context.Context) error {
	r := c.runtime
	if r.store == nil {
		return nil
	}
	leases, err := r.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		r.logger.Warn("list proxy lease cleanup facts failed", "error", err)
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
			r.logger.Warn("cleanup proxy lease fact failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
			cleanupErrors = append(cleanupErrors, fmt.Errorf("cleanup lease fact %q: %w", lease.GetLeaseId(), err))
		}
	}
	return errors.Join(cleanupErrors...)
}

func (c leaseCoordinator) cleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
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
			if strings.TrimSpace(current.GetProviderAccountId()) != "" {
				if err := r.leaseLocks.WithProviderAccountLock(ctx, current.GetProviderAccountId(), releaseProvider); err != nil {
					return err
				}
			} else if err := releaseProvider(ctx); err != nil {
				return err
			}
			leaseapp.ClearCleanupPending(current, false, true)
		}
		if !leaseapp.CleanupPending(current) {
			switch leaseapp.CleanupFinalStatus(current) {
			case leaseCleanupFinalExpired:
				return c.saveLeaseExpired(ctx, current)
			case leaseCleanupFinalReleased:
				return c.saveLeaseReleased(ctx, current)
			}
			return r.store.SaveLeaseFact(ctx, current)
		}
		return r.store.SaveLeaseFact(ctx, current)
	})
}
