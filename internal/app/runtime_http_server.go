package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	providerapp "github.com/byte-v-forge/proxy-runtime/internal/app/provider/application"
	checkapp "github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck/application"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings/application"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/gin-gonic/gin"
)

func (r *Runtime) serveHTTP(ctx context.Context, errCh chan<- error) {
	server := &http.Server{
		Addr:              r.cfg.RuntimeAddr,
		Handler:           r.httpHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	r.logger.Info("proxy-runtime http listening", "addr", r.cfg.RuntimeAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- fmt.Errorf("serve proxy-runtime http: %w", err)
	}
}

func (r *Runtime) httpHandler() http.Handler {
	return newRuntimeHTTPAPI(runtimeHTTPAPIDependencies{
		Leases:           r.leases,
		Settings:         newRuntimeSettingsApplication(runtimeSettingsDependencies(r)),
		Checks:           checkapp.New(runtimeCheckDependencies(r)),
		Providers:        providerapp.New(runtimeProviderDependencies(r)),
		Status:           newRuntimeStatusApplication(runtimeStatusDependencies(r)),
		MetricsUI:        newRuntimeMetricsApplication(runtimeMetricsDependencies(r)),
		Metrics:          r.metrics,
		MihomoAPIAddr:    r.cfg.Mihomo.APIAddr,
		ControlAuthToken: r.cfg.ControlAuthToken,
		ServiceAuthToken: r.cfg.ServiceAuthToken,
		Ready:            r.dataPlaneReadyFunc(),
		Logger:           r.logger,
		Clock:            r.clock,
	}).handler()
}

func (r *Runtime) dataPlaneReadyFunc() runtimeReadyFunc {
	return func() (bool, string) {
		reconcile := r.currentReconcileState()
		if reconcile.running {
			return false, "data plane reconcile running"
		}
		if reconcile.pending {
			return false, "data plane reconcile pending"
		}
		status := r.dataPlane.Status()
		if !status.Running {
			return false, appcore.FirstNonEmpty(status.LastError, "data plane is not running")
		}
		if status.LastError != "" {
			return false, "data plane reconcile failed"
		}
		if status.DesiredConfigHash != "" && status.DesiredConfigHash != status.AppliedConfigHash {
			return false, "data plane config projection is stale"
		}
		return true, ""
	}
}

type runtimeReadyFunc func() (bool, string)

type runtimeHTTPAPIDependencies struct {
	Leases           runtimeLeaseApplication
	Settings         settingsapp.Application
	Checks           checkapp.Service
	Providers        providerapp.Service
	Status           runtimeStatusApplication
	MetricsUI        runtimeMetricsApplication
	Metrics          *runtimeMetrics
	MihomoAPIAddr    string
	ControlAuthToken string
	ServiceAuthToken string
	Ready            runtimeReadyFunc
	Logger           *slog.Logger
	Clock            clock.Clock
}

type runtimeHTTPAPI struct {
	leases           runtimeLeaseApplication
	settings         settingsapp.Application
	checks           checkapp.Service
	providers        providerapp.Service
	status           runtimeStatusApplication
	metricsUI        runtimeMetricsApplication
	metrics          *runtimeMetrics
	auth             authapp.Application
	ready            runtimeReadyFunc
	logger           *slog.Logger
	clock            clock.Clock
	dashboardProxies dashboardProxyHandlers
}

func newRuntimeHTTPAPI(deps runtimeHTTPAPIDependencies) *runtimeHTTPAPI {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	trimmedAuthToken := strings.TrimSpace(deps.ControlAuthToken)
	api := &runtimeHTTPAPI{
		leases:    deps.Leases,
		settings:  deps.Settings,
		checks:    deps.Checks,
		providers: deps.Providers,
		status:    deps.Status,
		metricsUI: deps.MetricsUI,
		metrics:   deps.Metrics,
		auth:      authapp.NewApplication(trimmedAuthToken, deps.ServiceAuthToken),
		ready:     deps.Ready,
		logger:    logger,
		clock:     deps.Clock,
	}
	api.dashboardProxies = newDashboardProxyHandlers(deps.MihomoAPIAddr, trimmedAuthToken)
	return api
}

func (api *runtimeHTTPAPI) observe(operation string, startedAt time.Time, err error) {
	if api != nil && api.metrics != nil {
		api.metrics.Observe(operation, startedAt, err)
	}
}

func (api *runtimeHTTPAPI) handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(api.ginMiddleware())
	api.registerPublicHTTPRoutes(router)
	api.registerControlPlaneHTTPRoutes(router)
	router.NoMethod(api.handleGinMethodNotAllowed)
	router.NoRoute(api.handleGinNoRoute)
	return router
}

func (api *runtimeHTTPAPI) handleDashboardEntry(ctx *gin.Context) {
	if api.redirectToLoginIfRequired(ctx) {
		return
	}
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/overview")
}

func (api *runtimeHTTPAPI) handleGinMethodNotAllowed(ctx *gin.Context) {
	writeHTTPError(ctx.Writer, errors.New("method not allowed"), http.StatusMethodNotAllowed)
}

func (api *runtimeHTTPAPI) handleGinNoRoute(ctx *gin.Context) {
	path := ctx.Request.URL.Path
	if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/mihomo") {
		writeHTTPError(ctx.Writer, errors.New("not found"), http.StatusNotFound)
		return
	}
	if api.redirectToLoginIfRequired(ctx) {
		return
	}
	api.dashboardProxies.fallback(ctx)
}

func (api *runtimeHTTPAPI) ginMiddleware() gin.HandlerFunc {
	return httpapi.Middleware(httpapi.MiddlewareOptions{
		Logger:    api.logger,
		Authorize: api.authorize,
		WritePanicError: func(w http.ResponseWriter) {
			writeHTTPError(w, appcore.InternalError("", nil), http.StatusInternalServerError)
		},
	})
}

func runtimeMetricsFromRuntime(runtime *Runtime) *runtimeMetrics {
	if runtime == nil {
		return nil
	}
	return runtime.metrics
}

func runtimeMetricsDependencies(runtime *Runtime) runtimeMetricsApplicationDependencies {
	return runtimeMetricsApplicationDependencies{
		Metrics: runtimeMetricsFromRuntime(runtime),
	}
}

func runtimeProviderDependencies(runtime *Runtime) providerapp.Dependencies {
	if runtime == nil {
		return providerapp.Dependencies{}
	}
	var locks leaseapp.LockManager
	if runtime.leaseLocks != nil {
		locks = runtime.leaseLocks
	}
	var providerDescriptors providerapp.DescriptorsFunc
	if runtime.accountProviders != nil {
		providerDescriptors = runtime.accountProviders.Descriptors
	}
	return providerapp.Dependencies{
		Store:               runtime.store,
		LoadSettings:        runtime.settings.Load,
		ProviderDescriptors: providerDescriptors,
		Locks:               locks,
		LeaseOperations: func() providerapp.LeaseOperations {
			return runtime.leases
		},
		Logger: runtime.logger,
	}
}

func runtimeCheckDependencies(runtime *Runtime) checkapp.Dependencies {
	if runtime == nil {
		return checkapp.Dependencies{}
	}
	return checkapp.Dependencies{
		LoadSettings:   runtime.settings.Load,
		CheckClient:    runtime.checkProxyHTTPClient,
		ProbeExitIP:    runtime.probeExitIP,
		LookupGeo:      runtime.lookupIPGeo,
		CheckFraud:     runtime.checkIPFraud,
		RunEdgeCanary:  runtime.runEdgeCanary,
		ExitCheckCache: runtime.exitCheckCache,
	}
}

func runtimeSettingsDependencies(runtime *Runtime) runtimeSettingsApplicationDependencies {
	if runtime == nil {
		return runtimeSettingsApplicationDependencies{}
	}
	settingsApply := newRuntimeSettingsApplyScheduler(runtime)
	mihomoNative := newRuntimeSettingsMihomoNativeAdapter(runtime)
	return runtimeSettingsApplicationDependencies{
		Logger:                     runtime.logger,
		Settings:                   runtime.settings,
		ProxyUsers:                 runtime.cfg.ProxyUsers,
		IPFraudProviderViews:       runtime.ipFraudProviders.ProviderDescriptors,
		IPGeoProviderViews:         runtime.ipGeoProviders.ProviderDescriptors,
		LoadMihomoNativeSettings:   mihomoNative.Load,
		UpdateMihomoNativeSettings: mihomoNative.Update,
		ScheduleApply:              settingsApply.Schedule,
	}
}

func runtimeStatusDependencies(runtime *Runtime) runtimeStatusApplicationDependencies {
	if runtime == nil {
		return runtimeStatusApplicationDependencies{}
	}
	return runtimeStatusApplicationDependencies{
		RuntimeStatus: runtime.runtimeStatus,
	}
}

func (api *runtimeHTTPAPI) handleAuthSession(ctx *gin.Context) {
	api.auth.WriteSessionResponse(ctx.Writer, api.sessionAuthenticated(ctx.Request))
}

func (api *runtimeHTTPAPI) handleAuthWebSocketToken(ctx *gin.Context) {
	if err := api.auth.WriteWebSocketTokenResponse(ctx.Writer, api.clock.Now()); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) handleAuthLogin(ctx *gin.Context) {
	login, err := authapp.ReadLoginRequest(ctx.Request, readRequestBody)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	decision, err := api.auth.DecideLogin(login)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusUnauthorized)
		return
	}
	if err := api.auth.WriteLoginDecision(ctx.Writer, ctx.Request, decision, api.clock.Now()); err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) handleAuthLogout(ctx *gin.Context) {
	api.auth.WriteLogoutResponse(ctx.Writer, ctx.Request, ctx.Query("redirect"))
}

func (api *runtimeHTTPAPI) handleAuthLoginPage(ctx *gin.Context) {
	if err := api.auth.WriteLoginPageResponse(ctx.Writer, ctx.Request, api.sessionAuthenticated(ctx.Request)); err != nil {
		api.logger.Warn("render login page failed", "error_type", appcore.ErrorLogType(err))
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) redirectToLoginIfRequired(ctx *gin.Context) bool {
	return api.auth.WriteLoginRedirectIfRequired(ctx.Writer, ctx.Request, "/ui", api.clock.Now())
}

func (api *runtimeHTTPAPI) sessionAuthenticated(req *http.Request) bool {
	return api.auth.SessionAuthenticated(req, api.clock.Now())
}
