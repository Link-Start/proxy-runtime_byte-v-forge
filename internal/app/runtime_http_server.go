package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
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
	return newRuntimeHTTPAPI(r.service(), r.cfg.Mihomo.APIAddr, r.cfg.ControlAuthToken, r.cfg.ServiceAuthToken, func() (bool, string) {
		reconcile := r.currentReconcileState()
		if reconcile.running {
			return false, "data plane reconcile running"
		}
		if reconcile.pending {
			return false, "data plane reconcile pending"
		}
		status := r.dataPlane.Status()
		if !status.Running {
			return false, firstNonEmpty(status.LastError, "data plane is not running")
		}
		if status.LastError != "" {
			return false, "data plane reconcile failed"
		}
		if status.DesiredConfigHash != "" && status.DesiredConfigHash != status.AppliedConfigHash {
			return false, "data plane config projection is stale"
		}
		return true, ""
	}, r.logger, r.clock).handler()
}

type runtimeReadyFunc func() (bool, string)

type runtimeHTTPAPI struct {
	service          *RuntimeService
	auth             authapp.Application
	ready            runtimeReadyFunc
	logger           *slog.Logger
	clock            clock.Clock
	dashboardProxies dashboardProxyHandlers
}

func newRuntimeHTTPAPI(service *RuntimeService, mihomoAPIAddr string, authToken string, serviceAuthToken string, ready runtimeReadyFunc, logger *slog.Logger, clk clock.Clock) *runtimeHTTPAPI {
	if logger == nil {
		logger = slog.Default()
	}
	trimmedAuthToken := strings.TrimSpace(authToken)
	api := &runtimeHTTPAPI{
		service: service,
		auth:    authapp.NewApplication(trimmedAuthToken, serviceAuthToken),
		ready:   ready,
		logger:  logger,
		clock:   clk,
	}
	api.dashboardProxies = newDashboardProxyHandlers(mihomoAPIAddr, trimmedAuthToken)
	return api
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
			writeHTTPError(w, internalError("", nil), http.StatusInternalServerError)
		},
	})
}
