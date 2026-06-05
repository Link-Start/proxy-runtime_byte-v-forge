package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/randx"
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
	return newRuntimeHTTPAPI(r.service(), r.cfg.Mihomo.APIAddr, func() (bool, string) {
		status := r.dataPlane.Status()
		if status.Running {
			return true, ""
		}
		return false, firstNonEmpty(status.LastError, "data plane is not running")
	}, r.logger).handler()
}

type runtimeReadyFunc func() (bool, string)

type runtimeHTTPAPI struct {
	service       *RuntimeService
	mihomoAPIAddr string
	ready         runtimeReadyFunc
	logger        *slog.Logger
}

func newRuntimeHTTPAPI(service *RuntimeService, mihomoAPIAddr string, ready runtimeReadyFunc, logger *slog.Logger) *runtimeHTTPAPI {
	if logger == nil {
		logger = slog.Default()
	}
	return &runtimeHTTPAPI{service: service, mihomoAPIAddr: mihomoAPIAddr, ready: ready, logger: logger}
}

func (api *runtimeHTTPAPI) handler() http.Handler {
	mux := http.NewServeMux()
	api.registerPublicHTTPRoutes(mux)
	api.registerControlPlaneHTTPRoutes(mux)
	return api.withMiddleware(mux)
}

func (api *runtimeHTTPAPI) registerPublicHTTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", api.handleHealth)
	mux.HandleFunc("/readyz", api.handleReady)
	mux.HandleFunc("/proxy-runtime", api.handleDashboardEntry)
	mux.HandleFunc("/proxy-runtime/", api.handleDashboardEntry)
}

func (api *runtimeHTTPAPI) handleDashboardEntry(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	http.Redirect(w, req, "/api/proxy-runtime/mihomo/dashboard", http.StatusFound)
}

func (api *runtimeHTTPAPI) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := httpRequestID(req)
		w.Header().Set("X-Request-Id", requestID)
		start := time.Now()
		recorder := &httpStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if recovered := recover(); recovered != nil {
				if !recorder.wrote {
					writeHTTPError(recorder, internalError("", nil), http.StatusInternalServerError)
				}
				api.logger.Error("proxy-runtime http panic", "request_id", requestID, "method", req.Method, "path", req.URL.Path, "error", recovered)
			}
			api.logger.Info("proxy-runtime http request", "request_id", requestID, "method", req.Method, "path", req.URL.Path, "status", recorder.status, "duration_ms", time.Since(start).Milliseconds())
		}()
		next.ServeHTTP(recorder, req)
	})
}

type httpStatusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *httpStatusRecorder) WriteHeader(status int) {
	if r.wrote {
		return
	}
	r.status = status
	r.wrote = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *httpStatusRecorder) Write(data []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(data)
}

func (r *httpStatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func httpRequestID(req *http.Request) string {
	for _, header := range []string{"X-Request-Id", "X-Request-ID", "X-Correlation-Id"} {
		if value := strings.TrimSpace(req.Header.Get(header)); value != "" {
			return value
		}
	}
	value, err := randx.Hex(8)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return value
}
