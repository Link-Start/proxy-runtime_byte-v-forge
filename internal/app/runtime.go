package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"golang.org/x/sync/singleflight"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

type Runtime struct {
	cfg                 config.Config
	provider            provider.PoolProvider
	accountProviders    *providerregistry.Registry
	ipFraudProviders    *ipfraud.Registry
	ipGeoProviders      *ipgeo.Registry
	dataPlane           dataplane.Driver
	store               *store.RuntimeStores
	leaseLocks          leaseRuntimeLocks
	providerConcurrency leaseapp.ProviderAccountConcurrencyLimiter
	leaseCoordinator    leaseCoordinator
	dynamicIPSelector   *dynamicIPSelector
	settings            *runtimeSettingsStore
	metrics             *runtimeMetrics
	appService          *RuntimeService
	logger              *slog.Logger
	providerHTTPClient  *http.Client
	clock               clock.Clock

	refreshMu       sync.Mutex
	reconcileMu     sync.RWMutex
	reconcileState  runtimeReconcileState
	leaseRestoreMu  sync.RWMutex
	leaseRestore    runtimeLeaseRestoreState
	leaseWorkerMu   sync.RWMutex
	leaseWorker     runtimeLeaseWorkerState
	settingsApplyMu sync.RWMutex
	settingsApply   runtimeSettingsApplyState
	fraudChecker    ipFraudCheckerCache
	geoCache        *proxycheck.IPGeoCache
	geoLookupSF     singleflight.Group
	exitCheckCache  proxyExitCheckCache

	dynamicProfileMu        sync.RWMutex
	dynamicProfilePoolNodes []provider.Node
	dynamicProfileUpdatedAt time.Time

	reconcileCh chan struct{}
}

type RuntimeDeps struct {
	Config              config.Config
	ProxyProvider       provider.PoolProvider
	AccountProviders    *providerregistry.Registry
	IPFraudProviders    *ipfraud.Registry
	IPGeoProviders      *ipgeo.Registry
	DataPlane           dataplane.Driver
	Store               *store.RuntimeStores
	LeaseLocks          leaseRuntimeLocks
	ProviderConcurrency leaseapp.ProviderAccountConcurrencyLimiter
	ProviderHTTPClient  *http.Client
	Logger              *slog.Logger
	Clock               clock.Clock
}

func NewRuntime(deps RuntimeDeps) (*Runtime, error) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	if deps.DataPlane == nil {
		return nil, fmt.Errorf("data plane driver is required")
	}
	if deps.ProviderHTTPClient == nil {
		return nil, fmt.Errorf("provider HTTP client is required")
	}
	clk := deps.Clock
	if clk == nil {
		clk = clock.SystemClock{}
	}
	runtime := &Runtime{
		cfg:                 deps.Config,
		provider:            deps.ProxyProvider,
		accountProviders:    deps.AccountProviders,
		ipFraudProviders:    deps.IPFraudProviders,
		ipGeoProviders:      deps.IPGeoProviders,
		dataPlane:           deps.DataPlane,
		store:               deps.Store,
		leaseLocks:          deps.LeaseLocks,
		providerConcurrency: deps.ProviderConcurrency,
		providerHTTPClient:  deps.ProviderHTTPClient,
		settings:            newRuntimeSettingsStore(deps.Store, deps.AccountProviders, deps.IPFraudProviders, deps.IPGeoProviders),
		metrics:             newRuntimeMetrics(),
		logger:              logger,
		clock:               clk,
		geoCache:            proxycheck.NewIPGeoCache(clk),
		reconcileCh:         make(chan struct{}, 1),
	}
	runtime.exitCheckCache.clock = clk
	runtime.dynamicIPSelector = newDynamicIPSelector(runtimeDynamicIPSelectorDependencies(runtime))
	runtime.leaseCoordinator = newLeaseCoordinator(runtimeLeaseCoordinatorDependencies(runtime))
	runtime.appService = NewRuntimeService(runtime)
	return runtime, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	defer r.dataPlane.Stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 1)
	r.requestReconcile()

	var wg sync.WaitGroup
	spawn := func(fn func(context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(ctx)
		}()
	}
	spawn(r.reconcileLoop)
	spawn(r.leaseExpiryLoop)
	spawn(r.restoreActiveLeasesInBackground)
	spawn(func(ctx context.Context) { r.serveHTTP(ctx, errCh) })

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errCh:
	}
	cancel()
	wg.Wait()
	return runErr
}

func (r *Runtime) requestReconcile() {
	r.markReconcilePending()
	select {
	case r.reconcileCh <- struct{}{}:
	default:
	}
}

func (r *Runtime) reconcileLoop(ctx context.Context) {
	var ticker *time.Ticker
	var tick <-chan time.Time
	if r.cfg.RefreshInterval > 0 {
		ticker = time.NewTicker(r.cfg.RefreshInterval)
		defer ticker.Stop()
		tick = ticker.C
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			r.reconcile(ctx)
		case <-r.reconcileCh:
			r.reconcile(ctx)
		}
	}
}

func (r *Runtime) reconcile(ctx context.Context) {
	if err := r.runReconcile(ctx); err != nil {
		r.logger.Warn("proxy runtime reconcile failed", "error_type", appcore.ErrorLogType(err))
	}
	for {
		select {
		case <-r.reconcileCh:
			if err := r.runReconcile(ctx); err != nil {
				r.logger.Warn("proxy runtime reconcile failed", "error_type", appcore.ErrorLogType(err))
			}
		default:
			return
		}
	}
}

func (r *Runtime) runReconcile(ctx context.Context) error {
	r.markReconcileStarted()
	settingsApplyStarted := r.markSettingsApplyStartedIfPending()
	settingsApplyStartedAt := time.Now()
	err := r.refresh(ctx)
	r.markReconcileFinished(err)
	if settingsApplyStarted {
		r.observeRuntimeOperation(runtimeMetricSettingsApply, settingsApplyStartedAt, err)
		r.markSettingsApplyFinished(err)
	}
	return err
}

func (r *Runtime) refresh(ctx context.Context) error {
	r.refreshMu.Lock()
	defer r.refreshMu.Unlock()
	if err := r.projectMihomoNativeSettings(ctx); err != nil {
		return err
	}
	providerStartedAt := time.Now()
	nodes, err := r.provider.Fetch(ctx)
	r.observeRuntimeOperation(runtimeMetricProviderFetchBase, providerStartedAt, err)
	if err != nil && r.cfg.Provider != config.ProviderNone {
		r.logger.Warn("base provider fetch failed", "error_type", appcore.ErrorLogType(err))
		nodes = nil
	}
	sourceCfg, err := r.dataPlaneConfig(ctx)
	if err != nil {
		return err
	}
	dynamicProfileNodes := len(sourceCfg.Pool)
	sourceCfg.Pool = append(sourceCfg.Pool, nodes...)
	applyStartedAt := time.Now()
	sourceNodes, err := r.dataPlane.ApplyDesiredConfig(ctx, sourceCfg)
	r.observeRuntimeOperation(runtimeMetricDataPlaneApplyDesiredConfig, applyStartedAt, err)
	if err != nil {
		return err
	}
	r.refreshDynamicProfileSelectionMetadata(ctx)
	r.logger.Info("proxy runtime base refreshed", "provider_nodes", len(nodes), "dynamic_profile_nodes", dynamicProfileNodes, "mihomo_nodes", len(sourceNodes), "data_plane", r.dataPlane.Name())
	return nil
}

func (r *Runtime) observeRuntimeOperation(operation string, startedAt time.Time, err error) {
	if r != nil && r.metrics != nil {
		r.metrics.Observe(operation, startedAt, err)
	}
}
