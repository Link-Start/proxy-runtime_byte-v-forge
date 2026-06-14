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
	return c.saveFinalLeaseState(ctx, lease, leaseapp.FinalLeaseStateExpired)
}

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return c.saveFinalLeaseState(ctx, lease, leaseapp.FinalLeaseStateReleased)
}

func (c leaseCoordinator) saveLeaseCleanupProgress(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	stage, err := leaseapp.SaveCleanupProgress(ctx, c.deps.store, c.deps.providerConcurrency, lease)
	return c.finalLeaseSaveError(lease, stage, err)
}

func (c leaseCoordinator) saveFinalLeaseState(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, state leaseapp.FinalLeaseState) error {
	stage, err := leaseapp.SaveFinalLeaseState(ctx, c.deps.store, c.deps.providerConcurrency, lease, state)
	return c.finalLeaseSaveError(lease, stage, err)
}

func (c leaseCoordinator) finalLeaseSaveError(lease *proxyruntimev1.ProxyDynamicLease, stage leaseapp.FinalLeaseSaveStage, err error) error {
	if err == nil {
		return nil
	}
	if stage == leaseapp.FinalLeaseSaveConcurrencyRelease {
		c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
		return nil
	}
	return err
}
