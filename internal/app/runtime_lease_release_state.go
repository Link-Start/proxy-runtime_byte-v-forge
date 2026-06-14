package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) saveLeaseReleaseCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	return leaseapp.SaveReleaseCleanupFailure(ctx, c.deps.store, lease, routePending, providerPending, message)
}

func (c leaseCoordinator) saveLeaseCleanupRetry(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, message string) error {
	return leaseapp.SaveCleanupRetry(ctx, c.deps.store, lease, message)
}

func (c leaseCoordinator) saveLeaseExpiredCleanupFailure(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	return leaseapp.SaveExpiredCleanupFailure(ctx, c.deps.store, lease, routePending, providerPending, message)
}

func (c leaseCoordinator) saveLeaseExpired(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if err := leaseapp.SaveExpired(ctx, c.deps.store, lease); err != nil {
		return err
	}
	if err := leaseapp.ReleaseLeaseConcurrencySlot(ctx, c.deps.providerConcurrency, lease); err != nil {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if err := leaseapp.SaveReleased(ctx, c.deps.store, lease); err != nil {
		return err
	}
	if err := leaseapp.ReleaseLeaseConcurrencySlot(ctx, c.deps.providerConcurrency, lease); err != nil {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
	}
	return nil
}
