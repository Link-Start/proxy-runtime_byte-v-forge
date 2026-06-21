package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/gin-gonic/gin"
)

type runtimeLeaseApplication struct {
	leases *leaseapp.Application
}

func newRuntimeLeaseApplication(deps leaseapp.Dependencies) runtimeLeaseApplication {
	return runtimeLeaseApplication{leases: leaseapp.NewApplication(deps)}
}

func runtimeLeaseDependencies(runtime *Runtime) leaseapp.Dependencies {
	if runtime == nil {
		return leaseapp.Dependencies{}
	}
	return leaseapp.Dependencies{
		Repository:  runtime.store,
		Coordinator: runtime.leaseCoordinator,
		Worker:      runtime.leaseCoordinator.workerProcessor(),
		Logger:      runtime.logger,
	}
}

func (a runtimeLeaseApplication) GetProxyDynamicLeaseFact(ctx context.Context, leaseID string) (*proxygatewayv1.ProxyDynamicLease, error) {
	return a.leases.Get(ctx, leaseID)
}

func (a runtimeLeaseApplication) ListProxyDynamicLeaseFacts(ctx context.Context, options leaseapp.ListOptions) (*proxygatewayv1.ListProxyDynamicLeasesResponse, error) {
	leases, err := a.leases.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.ListProxyDynamicLeasesResponse{Leases: leases}, nil
}

func (a runtimeLeaseApplication) AcquireProxyLease(ctx context.Context, advertisedHost string, req *proxygatewayv1.AcquireProxyLeaseRequest) (*proxygatewayv1.AcquireProxyLeaseResponse, error) {
	return a.leases.Acquire(ctx, advertisedHost, req)
}

func (a runtimeLeaseApplication) ReleaseProxyLease(ctx context.Context, req *proxygatewayv1.ReleaseProxyLeaseRequest) (*proxygatewayv1.ReleaseProxyLeaseResponse, error) {
	return a.leases.Release(ctx, req)
}

func (a runtimeLeaseApplication) RestoreActiveLeases(ctx context.Context) error {
	return a.leases.RestoreActive(ctx)
}

func (a runtimeLeaseApplication) ExpireDueLeaseFacts(ctx context.Context) error {
	return a.leases.ExpireDue(ctx)
}

func (a runtimeLeaseApplication) CleanupPendingLeaseFacts(ctx context.Context) error {
	return a.leases.CleanupPending(ctx)
}

func (a runtimeLeaseApplication) CleanupPendingLeaseFact(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	return a.leases.Cleanup(ctx, lease)
}

func runtimeLeaseCoordinatorDependencies(runtime *Runtime) leaseCoordinatorDependencies {
	if runtime == nil {
		return leaseCoordinatorDependencies{}
	}
	var store leaseapp.OrchestrationStore
	if runtime.store != nil {
		store = runtime.store
	}
	var locks leaseapp.LockManager
	if runtime.leaseLocks != nil {
		locks = runtime.leaseLocks
	}
	return leaseCoordinatorDependencies{
		cfg:                     runtime.cfg,
		store:                   store,
		settings:                runtime.settings,
		clock:                   runtime.clock,
		locks:                   locks,
		dataPlane:               leaseRuntimeDataPlaneApplier{dataPlane: runtime.dataPlane, metrics: runtime.metrics},
		dynamicIPSelector:       runtime.dynamicIPSelector,
		sessionProviders:        leaseRegistrySessionProviderFactory{registry: runtime.accountProviders, client: runtime.providerHTTPClient, metrics: runtime.metrics, clock: runtime.clock},
		providerConcurrency:     runtime.providerConcurrency,
		logger:                  runtime.logger,
		exitCheckCache:          runtime.exitCheckCache,
		leaseListener:           runtime.leaseListener,
		localListenerEndpoint:   runtime.localListenerEndpoint,
		sessionAdvertisedHost:   runtime.sessionAdvertisedHost,
		dynamicLeaseDialerProxy: runtime.dynamicLeaseDialerProxy,
		closeInUserConnections:  runtime.closeMihomoInUserConnections,
	}
}

func (api *runtimeHTTPAPI) handleLeases(ctx *gin.Context) {
	options, err := parseLeaseListOptions(ctx)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	startedAt := time.Now()
	response, err := api.leases.ListProxyDynamicLeaseFacts(ctx.Request.Context(), options)
	api.observe(runtimeMetricLeaseList, startedAt, err)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleLease(ctx *gin.Context) {
	leaseID := strings.TrimSpace(ctx.Param("lease_id"))
	lease, err := api.leases.GetProxyDynamicLeaseFact(ctx.Request.Context(), leaseID)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, lease)
}

func (api *runtimeHTTPAPI) handleAcquireLease(ctx *gin.Context) {
	var body proxygatewayv1.AcquireProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	startedAt := time.Now()
	response, err := api.leases.AcquireProxyLease(ctx.Request.Context(), advertisedProxyHost(ctx.Request), &body)
	api.observe(runtimeMetricLeaseAcquire, startedAt, err)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func parseLeaseListOptions(ctx *gin.Context) (leaseapp.ListOptions, error) {
	options, err := leaseapp.ParseListOptions(ctx.Request.URL.Query())
	if err != nil {
		return leaseapp.ListOptions{}, appcore.InvalidArgument(err.Error(), err)
	}
	return options, nil
}

func (api *runtimeHTTPAPI) handleReleaseLease(ctx *gin.Context) {
	var body proxygatewayv1.ReleaseProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	startedAt := time.Now()
	response, err := api.leases.ReleaseProxyLease(ctx.Request.Context(), &body)
	api.observe(runtimeMetricLeaseRelease, startedAt, err)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

const (
	leaseExpirySweepInterval   = 30 * time.Second
	leaseCleanupAttemptTimeout = 20 * time.Second
)

type leaseWorkerTask func(context.Context) error

func (r *Runtime) leaseExpiryLoop(ctx context.Context) {
	r.runLeaseExpirySweep(ctx)
	ticker := time.NewTicker(leaseExpirySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runLeaseExpirySweep(ctx)
		}
	}
}

func (r *Runtime) runLeaseExpirySweep(ctx context.Context) {
	r.markLeaseWorkerStarted()
	err := errors.Join(
		r.runLeaseWorkerTask(ctx, runtimeMetricLeaseWorkerExpireDue, "expire proxy leases", leaseCleanupAttemptTimeout, r.leases.ExpireDueLeaseFacts),
		r.runLeaseWorkerTask(ctx, runtimeMetricLeaseWorkerCleanupPending, "cleanup pending proxy leases", leaseCleanupAttemptTimeout, r.leases.CleanupPendingLeaseFacts),
	)
	r.markLeaseWorkerFinished(err)
}

func (r *Runtime) runLeaseWorkerTask(ctx context.Context, operation string, name string, timeout time.Duration, task leaseWorkerTask) error {
	if task == nil {
		return nil
	}
	startedAt := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := task(taskCtx)
	r.observeRuntimeOperation(operation, startedAt, err)
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	r.logger.Warn(name+" failed", "error_type", appcore.ErrorLogType(err), "duration_ms", time.Since(startedAt).Milliseconds())
	return err
}
