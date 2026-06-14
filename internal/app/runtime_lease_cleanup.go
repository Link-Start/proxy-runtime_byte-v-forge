package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFacts(ctx context.Context) error {
	if c.deps.store == nil {
		return nil
	}
	leases, err := c.deps.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		c.warn("list proxy lease cleanup facts failed", "error", err)
		return err
	}
	return leaseapp.ProcessLeaseBatch(ctx, leaseapp.BatchInput{
		Leases:      leases,
		Timeout:     leaseCleanupAttemptTimeout,
		ErrorPrefix: "cleanup lease fact",
		Process:     c.cleanupPendingLeaseFact,
		Observe: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("cleanup proxy lease fact failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
	})
}
