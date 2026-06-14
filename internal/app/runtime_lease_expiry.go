package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) expireDueLeaseFacts(ctx context.Context) error {
	return leaseapp.ProcessExpiredActiveFacts(ctx, leaseapp.WorkerBatchInput{
		Store:   c.deps.store,
		Timeout: leaseCleanupAttemptTimeout,
		Process: c.expireLeaseFact,
		Observe: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("expire proxy lease failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
	})
}
