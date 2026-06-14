package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) cleanupPendingLeaseFacts(ctx context.Context) error {
	return leaseapp.ProcessCleanupPendingFacts(ctx, leaseapp.WorkerBatchInput{
		Store:   c.deps.store,
		Timeout: leaseCleanupAttemptTimeout,
		Process: c.cleanupPendingLeaseFact,
		Observe: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("cleanup proxy lease fact failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
		ObserveList: func(err error) {
			c.warn("list proxy lease cleanup facts failed", "error", err)
		},
	})
}
