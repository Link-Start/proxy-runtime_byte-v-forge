package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

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
		PlaygroundAccountID: playgroundProfileID,
		PlaygroundUsername:  playgroundUsername,
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
	selector := leaseDynamicIPSelectionAdapter{selector: f.deps.dynamicIPSelector}
	selectedRunnerFactory := leaseSelectedAcquireAttemptRunnerFactory{
		deps:           f.deps,
		settings:       settings,
		advertisedHost: f.advertisedHost,
		request:        f.request,
	}
	return leaseapp.DynamicAcquireAttemptRunner{
		Select:            selector.Select,
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
