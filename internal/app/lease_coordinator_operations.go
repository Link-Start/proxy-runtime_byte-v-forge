package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (c leaseCoordinator) Acquire(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	return c.acquireLease(ctx, advertisedHost, req)
}

func (c leaseCoordinator) Release(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	return c.releaseLease(ctx, req)
}

func (c leaseCoordinator) RestoreActive(ctx context.Context) error {
	return c.restoreActiveLeases(ctx)
}

func (c leaseCoordinator) ExpireDue(ctx context.Context) error {
	return c.expireDueLeaseFacts(ctx)
}

func (c leaseCoordinator) CleanupPending(ctx context.Context) error {
	return c.cleanupPendingLeaseFacts(ctx)
}

func (c leaseCoordinator) Cleanup(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return c.cleanupPendingLeaseFact(ctx, lease)
}
