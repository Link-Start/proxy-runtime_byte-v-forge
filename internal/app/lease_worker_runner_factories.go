package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

type leaseWorkerProcessorFactory struct {
	deps    leaseCoordinatorDependencies
	restore leaseapp.RestoreLeaseRouteRunner
	expire  leaseapp.ExpireLeaseRunner
	cleanup leaseapp.CleanupPendingLeaseRunner
}

func (f leaseWorkerProcessorFactory) New() leaseapp.WorkerProcessor {
	return leaseapp.WorkerProcessor{
		Store:             f.deps.store,
		Clock:             f.deps.clock,
		RestoreTimeout:    leaseRestoreRouteTimeout,
		CleanupTimeout:    leaseCleanupAttemptTimeout,
		Restore:           f.restore.Restore,
		Expire:            f.expire.Expire,
		CleanupPendingOne: f.cleanup.Cleanup,
		ObserveRestore:    f.observeRestore,
		ObserveRestoreList: func(err error) {
			f.warn("list proxy leases for restore failed", "error_type", appcore.ErrorLogType(err))
		},
		ObserveExpire: f.observeExpire,
		ObserveCleanup: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			f.warn("cleanup proxy lease fact failed", leaseWorkerObserverFields(lease, err)...)
		},
		ObserveCleanupList: func(err error) {
			f.warn("list proxy lease cleanup facts failed", "error_type", appcore.ErrorLogType(err))
		},
	}
}

func (f leaseWorkerProcessorFactory) observeRestore(lease *proxyruntimev1.ProxyDynamicLease, err error) {
	f.warn("restore proxy lease route failed", leaseWorkerObserverFields(lease, err)...)
}

func (f leaseWorkerProcessorFactory) observeExpire(lease *proxyruntimev1.ProxyDynamicLease, err error) {
	f.warn("expire proxy lease failed", leaseWorkerObserverFields(lease, err)...)
}

func (f leaseWorkerProcessorFactory) warn(message string, args ...any) {
	if f.deps.logger != nil {
		f.deps.logger.Warn(message, args...)
	}
}

func leaseWorkerObserverFields(lease *proxyruntimev1.ProxyDynamicLease, err error) []any {
	if lease == nil {
		return []any{"error_type", appcore.ErrorLogType(err)}
	}
	return []any{
		leaseapp.LabelLeaseID, lease.GetLeaseId(),
		leaseapp.LabelAccountID, lease.GetAccountId(),
		leaseapp.LabelProviderAccountID, lease.GetProviderAccountId(),
		"error_type", appcore.ErrorLogType(err),
	}
}

type leaseRestoreRunnerFactory struct {
	deps leaseCoordinatorDependencies
}

func (f leaseRestoreRunnerFactory) New() leaseapp.RestoreLeaseRouteRunner {
	return leaseapp.RestoreLeaseRouteRunner{ResolveRestorer: f.resolveRestorer}
}

func (f leaseRestoreRunnerFactory) resolveRestorer(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) (leaseapp.LeaseRouteRestorer, error) {
	settings, err := f.deps.settings.load(ctx)
	if err != nil {
		return leaseapp.LeaseRouteRestorer{}, err
	}
	return leaseRouteRestorerFactory{
		deps:     f.deps,
		settings: settings,
		adapter:  leaseSettingsAdapterFactory{deps: f.deps}.New(),
	}.New(), nil
}

type leaseExpireRunnerFactory struct {
	deps     leaseCoordinatorDependencies
	settings leaseapp.SettingsAdapter[*runtimeSettingsFile]
}

func (f leaseExpireRunnerFactory) New() leaseapp.ExpireLeaseRunner {
	return leaseapp.ExpireLeaseRunner{
		Store:                             f.deps.store,
		Limiter:                           f.deps.providerConcurrency,
		Locks:                             f.deps.locks,
		DataPlane:                         f.deps.dataPlane,
		Factory:                           f.deps.sessionProviders,
		Clock:                             f.deps.clock,
		LocalProtocol:                     f.deps.cfg.LocalProtocol,
		IsNotFound:                        store.IsNotFound,
		ResolveGatewaysForLease:           f.settings.ProviderGatewaysResolver,
		ObserveProviderReleaseFailure:     warnLeaseProviderSessionReleaseFailed(f.deps.logger),
		ObserveFinalConcurrencyReleaseErr: warnFinalLeaseConcurrencyReleaseFailed(f.deps.logger),
	}
}

type leaseCleanupPendingRunnerFactory struct {
	deps     leaseCoordinatorDependencies
	settings leaseapp.SettingsAdapter[*runtimeSettingsFile]
}

func (f leaseCleanupPendingRunnerFactory) New() leaseapp.CleanupPendingLeaseRunner {
	return leaseapp.CleanupPendingLeaseRunner{
		Store:                             f.deps.store,
		Limiter:                           f.deps.providerConcurrency,
		Locks:                             f.deps.locks,
		DataPlane:                         f.deps.dataPlane,
		Factory:                           f.deps.sessionProviders,
		LocalProtocol:                     f.deps.cfg.LocalProtocol,
		IsNotFound:                        store.IsNotFound,
		ResolveGatewaysForLease:           f.settings.ProviderGatewaysResolver,
		ObserveProviderReleaseFailure:     warnLeaseProviderSessionReleaseFailed(f.deps.logger),
		ObserveFinalConcurrencyReleaseErr: warnFinalLeaseConcurrencyReleaseFailed(f.deps.logger),
	}
}

func warnFinalLeaseConcurrencyReleaseFailed(logger leaseapp.Logger) leaseapp.LeaseObserver {
	return func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
		_ = ctx
		if logger == nil || lease == nil {
			return
		}
		logger.Warn("release provider account concurrency slot failed", leaseapp.LabelLeaseID, lease.GetLeaseId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
	}
}
