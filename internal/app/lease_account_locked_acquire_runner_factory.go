package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
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
