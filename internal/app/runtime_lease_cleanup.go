package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	leaseCleanupRoutePendingLabel    = "route_cleanup_pending"
	leaseCleanupProviderPendingLabel = "provider_cleanup_pending"
	leaseCleanupFinalStatusLabel     = "cleanup_final_status"
	leaseCleanupFinalFailed          = "failed"
	leaseCleanupFinalExpired         = "expired"
	leaseCleanupFinalReleased        = "released"
)

func markLeaseCleanupPending(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, finalStatus string) {
	if lease == nil {
		return
	}
	session := lease.GetSession()
	if session == nil {
		session = &proxyruntimev1.ProxySession{}
		lease.Session = session
	}
	if session.Labels == nil {
		session.Labels = map[string]string{}
	}
	if routePending {
		session.Labels[leaseCleanupRoutePendingLabel] = "true"
	}
	if providerPending {
		session.Labels[leaseCleanupProviderPendingLabel] = "true"
	}
	if strings.TrimSpace(finalStatus) != "" {
		session.Labels[leaseCleanupFinalStatusLabel] = strings.TrimSpace(finalStatus)
	}
}

func clearLeaseCleanupPending(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool) {
	if lease == nil || lease.GetSession() == nil || lease.GetSession().Labels == nil {
		return
	}
	if routePending {
		delete(lease.GetSession().Labels, leaseCleanupRoutePendingLabel)
	}
	if providerPending {
		delete(lease.GetSession().Labels, leaseCleanupProviderPendingLabel)
	}
}

func leaseCleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return leaseRouteCleanupPending(lease) || leaseProviderCleanupPending(lease)
}

func leaseRouteCleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return leaseCleanupLabel(lease, leaseCleanupRoutePendingLabel) == "true"
}

func leaseProviderCleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return leaseCleanupLabel(lease, leaseCleanupProviderPendingLabel) == "true"
}

func leaseCleanupFinalStatus(lease *proxyruntimev1.ProxyDynamicLease) string {
	return leaseCleanupLabel(lease, leaseCleanupFinalStatusLabel)
}

func leaseCleanupLabel(lease *proxyruntimev1.ProxyDynamicLease, key string) string {
	if lease == nil || lease.GetSession() == nil {
		return ""
	}
	return strings.TrimSpace(lease.GetSession().GetLabels()[key])
}

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
		if err := c.cleanupPendingLeaseFact(ctx, lease); err != nil {
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
		if !leaseCleanupPending(current) {
			return nil
		}
		if leaseRouteCleanupPending(current) {
			if err := c.deleteLeaseRoute(ctx, current); err != nil {
				_ = c.saveLeaseCleanupRetry(ctx, current, "lease route cleanup failed")
				return err
			}
			clearLeaseCleanupPending(current, true, false)
		}
		if leaseProviderCleanupPending(current) {
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
			clearLeaseCleanupPending(current, false, true)
		}
		if !leaseCleanupPending(current) {
			switch leaseCleanupFinalStatus(current) {
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
