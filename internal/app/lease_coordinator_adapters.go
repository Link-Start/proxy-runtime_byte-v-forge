package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

func (c leaseCoordinator) leaseRouteRetirer() leaseapp.LeaseRouteRetirer {
	settings := c.settingsAdapter()
	sideEffects := c.routeSideEffects()
	return leaseapp.LeaseRouteRetirer{
		Store:                   c.deps.store,
		Limiter:                 c.deps.providerConcurrency,
		Locks:                   c.deps.locks,
		DataPlane:               c.deps.dataPlane,
		Factory:                 c.deps.sessionProviders,
		LocalProtocol:           c.deps.cfg.LocalProtocol,
		ResolveGatewaysForLease: settings.ProviderGatewaysResolver,
		AfterRouteCleanup: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			sideEffects.afterRouteChange(ctx, lease.GetAccountId())
		},
		ObserveProviderReleaseFailure: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
			_ = ctx
			c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		},
		ObserveFinalConcurrencyReleaseErr: c.warnFinalConcurrencyReleaseFailed,
	}
}

func (c leaseCoordinator) releaseRunner() leaseapp.ReleaseRunner {
	retirer := c.leaseRouteRetirer()
	return leaseapp.ReleaseRunner{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		IsNotFound: isStoreNotFound,
		Retire:     retirer.Retire,
	}
}

func (c leaseCoordinator) leaseRouteRestorer(settings *runtimeSettingsFile) leaseapp.LeaseRouteRestorer {
	settingsAdapter := c.settingsAdapter()
	return leaseapp.LeaseRouteRestorer{
		Limiter:            c.deps.providerConcurrency,
		Store:              c.deps.store,
		DataPlane:          c.deps.dataPlane,
		Factory:            c.deps.sessionProviders,
		DefaultTTL:         leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:          providerAccountConcurrencyTTLBuffer,
		SlotReleaseTimeout: leaseRestoreSlotReleaseTimeout,
		LocalProtocol:      c.deps.cfg.LocalProtocol,
		Limit: func(lease *proxyruntimev1.ProxyDynamicLease) uint32 {
			return dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), leaseapp.ConcurrencyPolicy(lease))
		},
		ResolveGateways: func(lease *proxyruntimev1.ProxyDynamicLease) leaseapp.ProviderSessionGatewaysResolver {
			return settingsAdapter.ProviderGatewaysResolverForSettings(settings, lease)
		},
		ResolveLineBinding: settingsAdapter.RouteLineBindingResolver(settings),
	}
}

func (c leaseCoordinator) acquiredRouteApplier(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.AcquiredRouteApplier {
	sideEffects := c.routeSideEffects()
	return leaseapp.AcquiredRouteApplier{
		Store:            c.deps.store,
		DataPlane:        c.deps.dataPlane,
		Clock:            c.deps.clock,
		LocalProtocol:    c.deps.cfg.LocalProtocol,
		Managed:          true,
		FallbackProtocol: "http",
		ResolveListener: func(ctx context.Context, accountID string, leaseID string) (leaseapp.Listener, error) {
			return c.deps.leaseListener(ctx, settings, accountID, leaseID)
		},
		ResolveEgress: func(ctx context.Context, listener leaseapp.Listener) (*proxyruntimev1.ProxyEndpoint, error) {
			_ = ctx
			return c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(advertisedHost, listener))
		},
		AfterApply: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) {
			sideEffects.afterRouteChange(ctx, req.GetAccountId())
		},
	}
}

func (c leaseCoordinator) providerAccountAcquireRunner(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, leaseID string, concurrencyHolder string) leaseapp.ProviderAccountAcquireRunner {
	applier := c.acquiredRouteApplier(settings, advertisedHost, req)
	settingsAdapter := c.settingsAdapter()
	return leaseapp.ProviderAccountAcquireRunner{
		Store:              c.deps.store,
		IDs:                c.deps.ids,
		Clock:              c.deps.clock,
		DataPlane:          c.deps.dataPlane,
		Logger:             c.deps.logger,
		Factory:            c.deps.sessionProviders,
		Locks:              c.deps.locks,
		ResolveLineBinding: settingsAdapter.RouteLineBindingResolver(settings),
		Apply: func(ctx context.Context, acquired leaseapp.ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
			lease, err := applier.Apply(ctx, leaseapp.AcquiredRouteApplierInput{
				Failure:           acquired.Failure,
				LeaseID:           leaseID,
				Request:           req,
				ProviderClient:    acquired.ProviderClient,
				ProviderAccountID: acquired.ProviderAccountID,
				ConcurrencyHolder: concurrencyHolder,
				Session:           acquired.Session,
				Nodes:             acquired.Nodes,
				DialerProxy:       acquired.DialerProxy,
				LineLabels:        acquired.LineLabels,
				SelectionPlan:     selectionPlan,
			})
			if err != nil {
				return nil, acquiredRouteApplyError(err)
			}
			return lease, nil
		},
	}
}

func (c leaseCoordinator) selectedAcquireAttemptRunner(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
	return leaseapp.SelectedAcquireAttemptRunner{
		Store:          c.deps.store,
		IDs:            c.deps.ids,
		Limiter:        c.deps.providerConcurrency,
		Locks:          c.deps.locks,
		DefaultTTL:     leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:      providerAccountConcurrencyTTLBuffer,
		ReleaseTimeout: leaseAcquireSlotReleaseTimeout,
		Limit: func(selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
			return dynamicProviderConcurrencyLimit(settings, leaseapp.SelectedDynamicProviderID(selectionPlan), policy)
		},
		Action: func(ctx context.Context, attempt leaseapp.SelectedAcquireAttempt) (*proxyruntimev1.ProxyDynamicLease, error) {
			runner := c.providerAccountAcquireRunner(settings, advertisedHost, req, selection.Plan, attempt.LeaseID, attempt.ConcurrencyHolder)
			lease, err := runner.Acquire(ctx, leaseapp.ProviderAccountAcquireRunInput{
				ProviderAccountID: attempt.ProviderAccountID,
				Gateway:           selection.Endpoint,
				Request:           req,
				SelectionPlan:     selection.Plan,
				ConcurrencyHolder: attempt.ConcurrencyHolder,
			})
			if err != nil {
				return nil, providerSessionAcquireError(err)
			}
			return lease, nil
		},
	}
}

func (c leaseCoordinator) accountLockedAcquireRunner(ctx context.Context, settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.AccountLockedAcquireRunner {
	retirer := c.leaseRouteRetirer()
	refresher := c.concurrencySlotRefreshRunner()
	attemptRunner := leaseapp.DynamicAcquireAttemptRunner{
		Select: func(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (leaseapp.DynamicIPSelection, error) {
			if c.deps.dynamicIPSelector == nil {
				return leaseapp.DynamicIPSelection{}, internalError("dynamic IP selector is not configured", nil)
			}
			return c.deps.dynamicIPSelector.selectDynamicIPEndpoint(ctx, req)
		},
		NewSelectedRunner: func(selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
			return c.selectedAcquireAttemptRunner(settings, advertisedHost, req, selection)
		},
		MapSelectionError: func(err error) error {
			return failedPrecondition("no dynamic IP endpoint candidate", err)
		},
		MapAttemptError: acquireAttemptSlotError,
	}
	return leaseapp.AccountLockedAcquireRunner{
		Store:               c.deps.store,
		Clock:               c.deps.clock,
		PlaygroundAccountID: playgroundProfileID,
		PlaygroundUsername:  playgroundUsername,
		Reuse:               refresher.Refresh,
		Replace:             retirer.Retire,
		RunAttempt: func(int) (*proxyruntimev1.ProxyDynamicLease, error) {
			return attemptRunner.Run(ctx, req, req.GetPolicy())
		},
		Retry: retryLeaseAcquireAttempt,
		Observe: func(attempt int, err error) {
			c.warn("dynamic IP lease attempt failed", leaseapp.LabelAccountID, req.GetAccountId(), leaseapp.LabelPurpose, req.GetPurpose(), "attempt", attempt, "error_type", errorLogType(err))
		},
	}
}

func (c leaseCoordinator) preparedAcquireRunner(advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.PreparedAcquireRunner {
	return leaseapp.PreparedAcquireRunner{
		Locks: c.deps.locks,
		Action: func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
			settings, err := c.deps.settings.load(ctx)
			if err != nil {
				return nil, err
			}
			runner := c.accountLockedAcquireRunner(ctx, settings, advertisedHost, req)
			lease, err := runner.Run(ctx, leaseapp.AccountLockedAcquireRunnerInput{
				Request:        req,
				EgressProfiles: settings.GetEgressProfiles(),
			})
			if err != nil && leaseapp.IsAcquirePolicyError(err) {
				return nil, leaseProfilePolicyError(err)
			}
			return lease, err
		},
	}
}
