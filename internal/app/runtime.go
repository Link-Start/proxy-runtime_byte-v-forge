package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

type Runtime struct {
	cfg                 config.Config
	provider            provider.PoolProvider
	accountProviders    *providerregistry.Registry
	ipFraudProviders    *ipfraud.Registry
	ipGeoProviders      *ipgeo.Registry
	dataPlane           dataplane.Driver
	store               *RuntimeStores
	leaseLocks          leaseRuntimeLocks
	providerConcurrency providerAccountConcurrencyLimiter
	leaseCoordinator    leaseCoordinator
	dynamicIPSelector   *dynamicIPSelector
	settings            *runtimeSettingsStore
	appService          *RuntimeService
	logger              *slog.Logger
	providerHTTPClient  *http.Client

	refreshMu      sync.Mutex
	reconcileMu    sync.RWMutex
	reconcileState runtimeReconcileState
	fraudChecker   ipFraudCheckerCache
	geoCache       ipGeoCache
	exitCheckCache proxyExitCheckCache

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
	Store               *RuntimeStores
	LeaseLocks          leaseRuntimeLocks
	ProviderConcurrency providerAccountConcurrencyLimiter
	ProviderHTTPClient  *http.Client
	Logger              *slog.Logger
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
		settings:            newRuntimeSettingsStore(deps.Store, deps.AccountProviders, deps.IPFraudProviders, deps.IPGeoProviders, logger),
		logger:              logger,
		reconcileCh:         make(chan struct{}, 1),
	}
	runtime.leaseCoordinator = newLeaseCoordinator(runtime)
	runtime.dynamicIPSelector = newDynamicIPSelector(runtime)
	runtime.appService = NewRuntimeService(runtime)
	return runtime, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	if err := r.refresh(ctx); err != nil {
		return err
	}
	defer r.dataPlane.Stop()
	if err := r.leaseCoordinator.restoreActiveLeases(ctx); err != nil {
		return err
	}
	errCh := make(chan error, 2)
	go r.reconcileLoop(ctx)
	go r.leaseExpiryLoop(ctx)
	go r.serveHTTP(ctx, errCh)
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
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
		r.logger.Warn("proxy runtime reconcile failed", "error", err)
	}
	for {
		select {
		case <-r.reconcileCh:
			if err := r.runReconcile(ctx); err != nil {
				r.logger.Warn("proxy runtime reconcile failed", "error", err)
			}
		default:
			return
		}
	}
}

func (r *Runtime) runReconcile(ctx context.Context) error {
	r.markReconcileStarted()
	err := r.refresh(ctx)
	r.markReconcileFinished(err)
	return err
}

func (r *Runtime) refresh(ctx context.Context) error {
	r.refreshMu.Lock()
	defer r.refreshMu.Unlock()
	if err := r.projectMihomoNativeSettings(ctx); err != nil {
		return err
	}
	nodes, err := r.provider.Fetch(ctx)
	if err != nil && r.cfg.Provider != config.ProviderNone {
		r.logger.Warn("base provider fetch failed", "error", err)
		nodes = nil
	}
	sourceCfg, err := r.dataPlaneConfig(ctx)
	if err != nil {
		return err
	}
	dynamicProfileNodes := len(sourceCfg.Pool)
	sourceCfg.Pool = append(sourceCfg.Pool, nodes...)
	sourceNodes, err := r.dataPlane.ApplyDesiredConfig(ctx, sourceCfg)
	if err != nil {
		return err
	}
	r.refreshDynamicProfileSelectionMetadata(ctx)
	r.logger.Info("proxy runtime base refreshed", "provider_nodes", len(nodes), "dynamic_profile_nodes", dynamicProfileNodes, "mihomo_nodes", len(sourceNodes), "data_plane", r.dataPlane.Name())
	return nil
}
