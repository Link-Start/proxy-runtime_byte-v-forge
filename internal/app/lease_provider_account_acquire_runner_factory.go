package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

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
	settings := f.settingsAdapter()
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

func (f leaseProviderAccountAcquireRunnerFactory) settingsAdapter() leaseapp.SettingsAdapter[*runtimeSettingsFile] {
	return leaseapp.SettingsAdapter[*runtimeSettingsFile]{
		ResolveLineBinding: func(ctx context.Context, settings *runtimeSettingsFile, accountID string) (string, map[string]string, error) {
			return f.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
		},
	}
}
