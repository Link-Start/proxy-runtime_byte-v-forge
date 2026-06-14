package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) saveLeaseReleaseCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	leaseapp.MarkReleaseCleanupFailure(lease, routePending, providerPending, message)
	return c.deps.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseCleanupRetry(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, message string) error {
	leaseapp.MarkCleanupRetry(lease, message)
	return c.deps.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseExpiredCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	leaseapp.MarkExpiredCleanupFailure(lease, routePending, providerPending, message)
	return c.deps.store.SaveLeaseFact(ctx, lease)
}

func (c leaseCoordinator) saveLeaseExpired(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	leaseapp.MarkExpired(lease)
	if err := c.deps.store.SaveLeaseFact(ctx, lease); err != nil {
		return err
	}
	if err := c.releaseLeaseConcurrencySlot(ctx, lease); err != nil {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	leaseapp.MarkReleased(lease)
	if err := c.deps.store.SaveLeaseFact(ctx, lease); err != nil {
		return err
	}
	if err := c.releaseLeaseConcurrencySlot(ctx, lease); err != nil {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}
