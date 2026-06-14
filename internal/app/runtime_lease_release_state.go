package app

import (
	"context"
	"errors"

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
	err := leaseapp.SaveExpiredFinalLeaseState(ctx, c.deps.store, c.deps.providerConcurrency, lease)
	return c.finalLeaseSaveError(lease, err)
}

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	err := leaseapp.SaveReleasedFinalLeaseState(ctx, c.deps.store, c.deps.providerConcurrency, lease)
	return c.finalLeaseSaveError(lease, err)
}

func (c leaseCoordinator) saveLeaseCleanupProgress(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	err := leaseapp.SaveCleanupProgress(ctx, c.deps.store, c.deps.providerConcurrency, lease)
	return c.finalLeaseSaveError(lease, err)
}

func (c leaseCoordinator) finalLeaseSaveError(lease *proxyruntimev1.ProxyDynamicLease, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, leaseapp.ErrFinalLeaseConcurrencyRelease) {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
		return nil
	}
	return err
}
