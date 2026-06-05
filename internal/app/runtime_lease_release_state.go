package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func (c leaseCoordinator) saveLeaseReleased(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	clearLeaseCleanupPending(lease, true, true)
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED
	lease.ErrorMessage = ""
	return r.store.SaveLeaseFact(ctx, lease)
}
