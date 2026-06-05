package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

type Runtime struct {
	cfg                    config.Config
	provider               provider.PoolProvider
	accountProviders       *providerregistry.Registry
	ipFraudProviders       *ipfraud.Registry
	dataPlane              dataplane.Driver
	store                  *PostgresStore
	leaseLocks             leaseRuntimeLocks
	leaseCoordinator       leaseCoordinator
	dynamicGatewaySelector *dynamicGatewaySelector
	nodeObservations       *proxyNodeObservationStore
	settings               *runtimeSettingsStore
	appService             *RuntimeService
	logger                 *slog.Logger

	poolSnapshot      poolSnapshotState
	sourceObservation sourceNodeObservation

	refreshMu      sync.Mutex
	reconcileMu    sync.RWMutex
	reconcileState runtimeReconcileState
	fraudChecker   ipFraudCheckerCache
	geoCache       ipGeoCache

	reconcileCh chan struct{}
}

func NewRuntime(cfg config.Config, proxyProvider provider.PoolProvider, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, dataPlane dataplane.Driver, store *PostgresStore, leaseLocks leaseRuntimeLocks, logger *slog.Logger) (*Runtime, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if dataPlane == nil {
		return nil, fmt.Errorf("data plane driver is required")
	}
	nodeObservations, err := newProxyNodeObservationStore(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	runtime := &Runtime{cfg: cfg, provider: proxyProvider, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, dataPlane: dataPlane, store: store, leaseLocks: leaseLocks, nodeObservations: nodeObservations, settings: newRuntimeSettingsStore(store, accountProviders, ipFraudProviders, logger), logger: logger, reconcileCh: make(chan struct{}, 1)}
	runtime.leaseCoordinator = newLeaseCoordinator(runtime)
	runtime.dynamicGatewaySelector = newDynamicGatewaySelector(runtime)
	runtime.appService = NewRuntimeService(runtime)
	return runtime, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	defer r.nodeObservations.Close()
	if err := r.refresh(ctx); err != nil {
		return err
	}
	defer r.dataPlane.Stop()
	errCh := make(chan error, 2)
	go r.reconcileLoop(ctx)
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
	nodes, err := r.provider.Fetch(ctx)
	if err != nil && r.cfg.Provider != config.ProviderNone {
		r.logger.Warn("base provider fetch failed", "error", err)
		nodes = nil
	}
	sourceCfg, err := r.dataPlaneConfig(ctx)
	if err != nil {
		return err
	}
	sourceCfg.Pool = nodes
	sourceCfg.Common = r.commonEgressService()
	sourceCfg.Local = r.defaultLocalService()
	sourceCfg.DynamicViaCommon = r.cfg.CommonEgressAddr != ""
	sourceNodes, err := r.dataPlane.ReconcileBase(ctx, sourceCfg)
	if err != nil {
		r.sourceObservation.recordError(err)
		return err
	}
	poolNodes := append(cloneNodes(nodes), cloneNodes(sourceNodes)...)
	r.refreshSourceNodeObservation(ctx)
	r.poolSnapshot.record(poolNodes)
	if err := r.leaseCoordinator.restoreActiveLeases(ctx); err != nil {
		return err
	}
	r.logger.Info("proxy runtime base refreshed", "pool_size", len(poolNodes), "data_plane", r.dataPlane.Name())
	return nil
}
