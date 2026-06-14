package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/random"
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
	return newRuntimeHTTPAPI(r.service(), r.cfg.Mihomo.APIAddr, r.cfg.ControlAuthToken, func() (bool, string) {
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
	}, r.logger).handler()
}

type runtimeReadyFunc func() (bool, string)

type runtimeHTTPAPI struct {
	service       *RuntimeService
	mihomoAPIAddr string
	authToken     string
	ready         runtimeReadyFunc
	logger        *slog.Logger
}

func newRuntimeHTTPAPI(service *RuntimeService, mihomoAPIAddr string, authToken string, ready runtimeReadyFunc, logger *slog.Logger) *runtimeHTTPAPI {
	if logger == nil {
		logger = slog.Default()
	}
	return &runtimeHTTPAPI{service: service, mihomoAPIAddr: mihomoAPIAddr, authToken: strings.TrimSpace(authToken), ready: ready, logger: logger}
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

func (api *runtimeHTTPAPI) registerPublicHTTPRoutes(router *gin.Engine) {
	router.GET("/healthz", api.handleHealth)
	router.GET("/readyz", api.handleReady)
	router.GET("/", api.handleDashboardEntry)
	router.HEAD("/", api.handleDashboardEntry)
}

func (api *runtimeHTTPAPI) handleDashboardEntry(ctx *gin.Context) {
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/proxies", "/mihomo/ui/#/setup?endpoint="+mihomoDashboardEndpointID)
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
	api.mihomoReverseProxy("/", "/ui/")(ctx)
}

func (api *runtimeHTTPAPI) ginMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := httpRequestID(ctx.Request)
		ctx.Header("X-Request-Id", requestID)
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				if !ctx.Writer.Written() {
					writeHTTPError(ctx.Writer, internalError("", nil), http.StatusInternalServerError)
				}
				ctx.Abort()
				api.logger.Error("proxy-runtime http panic", "request_id", requestID, "method", ctx.Request.Method, "path", ctx.Request.URL.Path, "error", recovered)
			}
			api.logger.Info("proxy-runtime http request", "request_id", requestID, "method", ctx.Request.Method, "path", ctx.Request.URL.Path, "status", ctx.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
		}()
		if !api.authorize(ctx) {
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func httpRequestID(req *http.Request) string {
	for _, header := range []string{"X-Request-Id", "X-Request-ID", "X-Correlation-Id"} {
		if value := strings.TrimSpace(req.Header.Get(header)); value != "" {
			return value
		}
	}
	value, err := random.Hex(8)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return value
}
