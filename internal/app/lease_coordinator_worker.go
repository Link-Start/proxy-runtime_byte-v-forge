package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) workerProcessor() leaseapp.WorkerProcessor {
	restorer := c.restoreLeaseRunner()
	expirer := c.expireLeaseRunner()
	cleaner := c.cleanupPendingLeaseRunner()
	return leaseapp.WorkerProcessor{
		Store:             c.deps.store,
		Clock:             c.deps.clock,
		RestoreTimeout:    leaseRestoreRouteTimeout,
		CleanupTimeout:    leaseCleanupAttemptTimeout,
		Restore:           restorer.Restore,
		Expire:            expirer.Expire,
		CleanupPendingOne: cleaner.Cleanup,
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

func (c leaseCoordinator) restoreLeaseRunner() leaseapp.RestoreLeaseRouteRunner {
	return leaseapp.RestoreLeaseRouteRunner{
		ResolveRestorer: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) (leaseapp.LeaseRouteRestorer, error) {
			settings, err := c.deps.settings.load(ctx)
			if err != nil {
				return leaseapp.LeaseRouteRestorer{}, err
			}
			return c.leaseRouteRestorer(settings), nil
		},
	}
}

func (c leaseCoordinator) expireLeaseRunner() leaseapp.ExpireLeaseRunner {
	settings := c.settingsAdapter()
	return leaseapp.ExpireLeaseRunner{
		Store:                             c.deps.store,
		Limiter:                           c.deps.providerConcurrency,
		Locks:                             c.deps.locks,
		DataPlane:                         c.deps.dataPlane,
		Factory:                           c.deps.sessionProviders,
		Clock:                             c.deps.clock,
		LocalProtocol:                     c.deps.cfg.LocalProtocol,
		IsNotFound:                        isStoreNotFound,
		ResolveGatewaysForLease:           settings.ProviderGatewaysResolver,
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	}
}

func (c leaseCoordinator) cleanupPendingLeaseRunner() leaseapp.CleanupPendingLeaseRunner {
	settings := c.settingsAdapter()
	return leaseapp.CleanupPendingLeaseRunner{
		Store:                             c.deps.store,
		Limiter:                           c.deps.providerConcurrency,
		Locks:                             c.deps.locks,
		DataPlane:                         c.deps.dataPlane,
		Factory:                           c.deps.sessionProviders,
		LocalProtocol:                     c.deps.cfg.LocalProtocol,
		IsNotFound:                        isStoreNotFound,
		ResolveGatewaysForLease:           settings.ProviderGatewaysResolver,
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	}
}
