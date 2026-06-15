package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) workerProcessor() leaseapp.WorkerProcessor {
	return leaseapp.WorkerProcessor{
		Store:             c.deps.store,
		Clock:             c.deps.clock,
		RestoreTimeout:    leaseRestoreRouteTimeout,
		CleanupTimeout:    leaseCleanupAttemptTimeout,
		Restore:           c.restoreLeaseRoute,
		Expire:            c.expireLeaseFact,
		CleanupPendingOne: c.cleanupPendingLeaseFact,
		ObserveRestore: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("restore proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
		},
		ObserveRestoreList: func(err error) {
			c.warn("list proxy leases for restore failed", "error", err)
		},
		ObserveExpire: func(lease *proxyruntimev1.ProxyDynamicLease, _ error) {
			c.warn("expire proxy lease failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
		ObserveCleanup: func(lease *proxyruntimev1.ProxyDynamicLease, _ error) {
			c.warn("cleanup proxy lease fact failed", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
		},
		ObserveCleanupList: func(err error) {
			c.warn("list proxy lease cleanup facts failed", "error", err)
		},
	}
}
