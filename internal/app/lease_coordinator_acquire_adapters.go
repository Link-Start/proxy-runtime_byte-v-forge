package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const leaseAcquireSlotReleaseTimeout = 5 * time.Second

func (c leaseCoordinator) providerAccountAcquireRunner(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, leaseID string, concurrencyHolder string) leaseapp.ProviderAccountAcquireRunner {
	applier := c.acquiredRouteApplier(settings, advertisedHost, req)
	apply := leaseapp.ProviderAccountAcquiredRouteApplier{
		Applier:           applier,
		LeaseID:           leaseID,
		Request:           req,
		SelectionPlan:     selectionPlan,
		ConcurrencyHolder: concurrencyHolder,
		MapError:          acquiredRouteApplyError,
	}
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
		Apply:              apply.Apply,
	}
}

func (c leaseCoordinator) selectedAcquireAttemptRunner(settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
	action := leaseapp.SelectedAttemptProviderAccountAction{
		Selection: selection,
		Request:   req,
		NewRunner: func(attempt leaseapp.SelectedAcquireAttempt) leaseapp.ProviderAccountAcquireRunner {
			return c.providerAccountAcquireRunner(settings, advertisedHost, req, selection.Plan, attempt.LeaseID, attempt.ConcurrencyHolder)
		},
		MapError: providerSessionAcquireError,
	}
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
		Action: action.Run,
	}
}

func (c leaseCoordinator) accountLockedAcquireRunner(ctx context.Context, settings *runtimeSettingsFile, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) leaseapp.AccountLockedAcquireRunner {
	retirer := c.leaseRouteRetirer()
	refresher := c.concurrencySlotRefreshRunner()
	selector := leaseDynamicIPSelectionAdapter{selector: c.deps.dynamicIPSelector}
	attemptRunner := leaseapp.DynamicAcquireAttemptRunner{
		Select: selector.Select,
		NewSelectedRunner: func(selection leaseapp.DynamicIPSelection) leaseapp.SelectedAcquireAttemptRunner {
			return c.selectedAcquireAttemptRunner(settings, advertisedHost, req, selection)
		},
		MapSelectionError: mapDynamicIPSelectionError,
		MapAttemptError:   acquireAttemptSlotError,
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
	action := leaseapp.SettingsPreparedAcquireAction[*runtimeSettingsFile]{
		Load:    c.deps.settings.load,
		Request: req,
		EgressProfiles: func(settings *runtimeSettingsFile) []*proxyruntimev1.EgressProfileSettings {
			return settings.GetEgressProfiles()
		},
		NewRunner: func(ctx context.Context, settings *runtimeSettingsFile) leaseapp.AccountLockedAcquireRunner {
			return c.accountLockedAcquireRunner(ctx, settings, advertisedHost, req)
		},
		MapPolicyError: leaseProfilePolicyError,
	}
	return leaseapp.PreparedAcquireRunner{
		Locks:  c.deps.locks,
		Action: action.Run,
	}
}
