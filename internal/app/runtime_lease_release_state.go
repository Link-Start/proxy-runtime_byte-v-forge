package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (c leaseCoordinator) saveLeaseReleaseCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	r := c.runtime
	markLeaseCleanupPending(lease, routePending, providerPending, leaseCleanupFinalReleased)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED
	lease.ErrorMessage = message
	return r.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseCleanupRetry(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, message string) error {
	r := c.runtime
	lease.ErrorMessage = message
	return r.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseExpiredCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	r := c.runtime
	markLeaseCleanupPending(lease, routePending, providerPending, leaseCleanupFinalExpired)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED
	lease.ErrorMessage = message
	return r.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseExpired(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	clearLeaseCleanupPending(lease, true, true)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED
	lease.ErrorMessage = ""
	if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
		return err
	}
	if err := r.releaseLeaseConcurrencySlot(ctx, lease); err != nil {
		r.logger.Warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	clearLeaseCleanupPending(lease, true, true)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED
	lease.ErrorMessage = ""
	if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
		return err
	}
	if err := r.releaseLeaseConcurrencySlot(ctx, lease); err != nil {
		r.logger.Warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}
