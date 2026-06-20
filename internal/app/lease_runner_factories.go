package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
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
	settings, err := f.deps.settings.Load(ctx)
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

type leaseAccountLockedAcquireRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	advertisedHost string
	request        *proxyruntimev1.AcquireProxyLeaseRequest
	retire         leaseapp.LeaseRouteRetirer
	reuse          leaseapp.RefreshConcurrencySlotRunner
}

func (f leaseAccountLockedAcquireRunnerFactory) New(ctx context.Context, settings *runtimeSettingsFile) leaseapp.AccountLockedAcquireRunner {
	attemptRunner := f.dynamicAttemptRunner(settings)
	return leaseapp.AccountLockedAcquireRunner{
		Store:               f.deps.store,
		Clock:               f.deps.clock,
		PlaygroundAccountID: kernel.PlaygroundProfileID,
		PlaygroundUsername:  kernel.PlaygroundUsername,
		Reuse:               f.reuse.Refresh,
		Replace:             f.retire.Retire,
		RunAttempt: func(int) (*proxyruntimev1.ProxyDynamicLease, error) {
			return attemptRunner.Run(ctx, f.request, f.request.GetPolicy())
		},
		Retry:   retryLeaseAcquireAttempt,
		Observe: f.observeAttemptFailure,
	}
}

func (f leaseAccountLockedAcquireRunnerFactory) dynamicAttemptRunner(settings *runtimeSettingsFile) leaseapp.DynamicAcquireAttemptRunner {
	selectedRunnerFactory := leaseSelectedAcquireAttemptRunnerFactory{
		deps:           f.deps,
		settings:       settings,
		advertisedHost: f.advertisedHost,
		request:        f.request,
	}
	return leaseapp.DynamicAcquireAttemptRunner{
		Select:            f.deps.dynamicIPSelector.SelectDynamicIPEndpoint,
		NewSelectedRunner: selectedRunnerFactory.New,
		MapSelectionError: mapDynamicIPSelectionError,
		MapAttemptError:   acquireAttemptSlotError,
	}
}

func (f leaseAccountLockedAcquireRunnerFactory) observeAttemptFailure(attempt int, err error) {
	if f.deps.logger == nil {
		return
	}
	f.deps.logger.Warn("dynamic IP lease attempt failed", leaseapp.LabelAccountID, f.request.GetAccountId(), leaseapp.LabelPurpose, f.request.GetPurpose(), "attempt", attempt, "error_type", appcore.ErrorLogType(err))
}

func mapDynamicIPSelectionError(err error) error {
	return appcore.FailedPrecondition("no dynamic IP endpoint candidate", err)
}

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

type leaseSelectedAcquireAttemptRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	settings       *runtimeSettingsFile
	advertisedHost string
	request        *proxyruntimev1.AcquireProxyLeaseRequest
}

func (f leaseSelectedAcquireAttemptRunnerFactory) New(selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
	providerRunnerFactory := leaseProviderAccountAcquireRunnerFactory{
		deps:           f.deps,
		settings:       f.settings,
		advertisedHost: f.advertisedHost,
		request:        f.request,
		selectionPlan:  selection.Plan,
	}
	action := leaseapp.SelectedAttemptProviderAccountAction{
		Selection: selection,
		Request:   f.request,
		NewRunner: providerRunnerFactory.New,
		MapError:  providerSessionAcquireError,
	}
	return leaseapp.SelectedAcquireAttemptRunner{
		Store:          f.deps.store,
		IDs:            f.deps.ids,
		Limiter:        f.deps.providerConcurrency,
		Locks:          f.deps.locks,
		DefaultTTL:     kernel.DefaultDynamicIPStickyTTL,
		TTLBuffer:      providerAccountConcurrencyTTLBuffer,
		ReleaseTimeout: leaseAcquireSlotReleaseTimeout,
		Limit:          f.limit,
		Action:         action.Run,
	}
}

func (f leaseSelectedAcquireAttemptRunnerFactory) limit(selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	return dynamicProviderConcurrencyLimit(f.settings, leaseapp.SelectedDynamicProviderID(selectionPlan), policy)
}

type leaseProviderAccountAcquireRunnerFactory struct {
	deps           leaseCoordinatorDependencies
	settings       *runtimeSettingsFile
	advertisedHost string
	request        *proxyruntimev1.AcquireProxyLeaseRequest
	selectionPlan  *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

func (f leaseProviderAccountAcquireRunnerFactory) New(attempt leaseapp.SelectedAcquireAttempt) leaseapp.ProviderAccountAcquireRunner {
	apply := leaseapp.ProviderAccountAcquiredRouteApplier{
		Applier:           f.acquiredRouteApplier(),
		LeaseID:           attempt.LeaseID,
		Request:           f.request,
		SelectionPlan:     f.selectionPlan,
		ConcurrencyHolder: attempt.ConcurrencyHolder,
		MapError:          acquiredRouteApplyError,
	}
	settings := leaseSettingsAdapterFactory{deps: f.deps}.New()
	return leaseapp.ProviderAccountAcquireRunner{
		Store:              f.deps.store,
		IDs:                f.deps.ids,
		Clock:              f.deps.clock,
		DataPlane:          f.deps.dataPlane,
		Logger:             f.deps.logger,
		Factory:            f.deps.sessionProviders,
		Locks:              f.deps.locks,
		ResolveLineBinding: settings.RouteLineBindingResolver(f.settings),
		Apply:              apply.Apply,
	}
}

func (f leaseProviderAccountAcquireRunnerFactory) acquiredRouteApplier() leaseapp.AcquiredRouteApplier {
	sideEffects := leaseRouteSideEffects{
		exitCheckCache:         f.deps.exitCheckCache,
		closeInUserConnections: f.deps.closeInUserConnections,
	}
	endpoint := leaseAcquiredEndpointAdapter{
		settings:              f.settings,
		advertisedHost:        f.advertisedHost,
		leaseListener:         f.deps.leaseListener,
		localListenerEndpoint: f.deps.localListenerEndpoint,
		sessionAdvertisedHost: f.deps.sessionAdvertisedHost,
	}
	return leaseapp.AcquiredRouteApplier{
		Store:            f.deps.store,
		DataPlane:        f.deps.dataPlane,
		Clock:            f.deps.clock,
		LocalProtocol:    f.deps.cfg.LocalProtocol,
		Managed:          true,
		FallbackProtocol: "http",
		ResolveListener:  endpoint.ResolveListener,
		ResolveEgress:    endpoint.ResolveEgress,
		AfterApply: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) {
			sideEffects.afterRouteChange(ctx, f.request.GetAccountId())
		},
	}
}
