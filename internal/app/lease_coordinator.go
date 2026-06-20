package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type leaseCoordinator struct {
	deps leaseCoordinatorDependencies
}

func newLeaseCoordinator(deps leaseCoordinatorDependencies) leaseCoordinator {
	if deps.clock == nil {
		deps.clock = clock.SystemClock{}
	}
	if deps.ids == nil {
		deps.ids = randomLeaseIDGenerator{byteLength: leaseIDByteLength}
	}
	return leaseCoordinator{deps: deps}
}

func (c leaseCoordinator) routeSideEffects() leaseRouteSideEffects {
	return leaseRouteSideEffects{
		exitCheckCache:         c.deps.exitCheckCache,
		closeInUserConnections: c.deps.closeInUserConnections,
	}
}

type leaseCoordinatorSettings interface {
	Load(context.Context) (*runtimeSettingsFile, error)
}

type leaseListenerFunc func(context.Context, *runtimeSettingsFile, string, string) (leaseapp.Listener, error)
type leaseEndpointFunc func(leaseapp.Listener, string) (*proxyruntimev1.ProxyEndpoint, error)
type leaseAdvertisedHostFunc func(string, leaseapp.Listener) string
type leaseDialerProxyFunc func(context.Context, *runtimeSettingsFile, string) (string, map[string]string, error)
type leaseConnectionCleanupFunc func(context.Context, []string)

type leaseCoordinatorDependencies struct {
	cfg                     config.Config
	store                   leaseapp.OrchestrationStore
	settings                leaseCoordinatorSettings
	clock                   clock.Clock
	ids                     leaseapp.IDGenerator
	locks                   leaseapp.LockManager
	dataPlane               leaseapp.DataPlaneApplier
	dynamicIPSelector       *dynamic.IPSelector
	sessionProviders        leaseapp.SessionProviderFactory
	providerConcurrency     leaseapp.ProviderAccountConcurrencyLimiter
	logger                  leaseapp.Logger
	exitCheckCache          *proxycheck.ExitCheckCache
	leaseListener           leaseListenerFunc
	localListenerEndpoint   leaseEndpointFunc
	sessionAdvertisedHost   leaseAdvertisedHostFunc
	dynamicLeaseDialerProxy leaseDialerProxyFunc
	closeInUserConnections  leaseConnectionCleanupFunc
}

func (c leaseCoordinator) Acquire(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	runner := c.preparedAcquireRunner(advertisedHost, req)
	lease, err := runner.Run(ctx, leaseapp.PreparedAcquireRunnerInput{
		Request: req,
	})
	if err != nil && leaseapp.IsAcquireRequestError(err) {
		return nil, appcore.InvalidArgument(err.Error(), err)
	}
	return lease, err
}

func (c leaseCoordinator) Release(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := c.releaseRunner().Release(ctx, req)
	if err != nil && leaseapp.IsReleaseLookupRequestError(err) {
		return nil, appcore.InvalidArgument(err.Error(), err)
	}
	return lease, err
}

func (c leaseCoordinator) leaseRouteRetirer() leaseapp.LeaseRouteRetirer {
	return leaseRouteRetirerFactory{
		deps:             c.deps,
		settings:         c.settingsAdapter(),
		sideEffects:      c.routeSideEffects(),
		observeFinalSlot: warnFinalLeaseConcurrencyReleaseFailed(c.deps.logger),
	}.New()
}

func (c leaseCoordinator) releaseRunner() leaseapp.ReleaseRunner {
	return leaseReleaseRunnerFactory{
		deps:   c.deps,
		retire: c.leaseRouteRetirer(),
	}.New()
}

func (c leaseCoordinator) leaseRouteRestorer(settings *runtimeSettingsFile) leaseapp.LeaseRouteRestorer {
	return leaseRouteRestorerFactory{
		deps:     c.deps,
		settings: settings,
		adapter:  c.settingsAdapter(),
	}.New()
}

func (c leaseCoordinator) workerProcessor() leaseapp.WorkerProcessor {
	return leaseWorkerProcessorFactory{
		deps:    c.deps,
		restore: c.restoreLeaseRunner(),
		expire:  c.expireLeaseRunner(),
		cleanup: c.cleanupPendingLeaseRunner(),
	}.New()
}

func (c leaseCoordinator) restoreLeaseRunner() leaseapp.RestoreLeaseRouteRunner {
	return leaseRestoreRunnerFactory{deps: c.deps}.New()
}

func (c leaseCoordinator) expireLeaseRunner() leaseapp.ExpireLeaseRunner {
	return leaseExpireRunnerFactory{
		deps:     c.deps,
		settings: c.settingsAdapter(),
	}.New()
}

func (c leaseCoordinator) cleanupPendingLeaseRunner() leaseapp.CleanupPendingLeaseRunner {
	return leaseCleanupPendingRunnerFactory{
		deps:     c.deps,
		settings: c.settingsAdapter(),
	}.New()
}
