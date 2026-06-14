package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireDueLeaseFacts(ctx context.Context) error {
	if c.deps.store == nil {
		return nil
	}
	leases, err := c.deps.store.ExpiredActiveLeaseFacts(ctx)
	if err != nil {
		return err
	}
	return leaseapp.ProcessLeaseBatch(ctx, leaseapp.BatchInput{
		Leases:      leases,
		Timeout:     leaseCleanupAttemptTimeout,
		ErrorPrefix: "expire lease fact",
		Process:     c.expireLeaseFact,
		Observe: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("expire proxy lease failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
	})
}
