package httpapi

import (
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

type Authorizer func(*gin.Context) bool

type PanicErrorWriter func(http.ResponseWriter)

type MiddlewareOptions struct {
	Logger          *slog.Logger
	Authorize       Authorizer
	WritePanicError PanicErrorWriter
}

func Middleware(opts MiddlewareOptions) gin.HandlerFunc {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx *gin.Context) {
		requestID := RequestID(ctx.Request)
		ctx.Header("X-Request-Id", requestID)
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				if !ctx.Writer.Written() {
					writePanicError(ctx.Writer, opts.WritePanicError)
				}
				ctx.Abort()
				logger.Error(
					"proxy-gateway http panic",
					"request_id", requestID,
					"method", ctx.Request.Method,
					"path", ctx.Request.URL.Path,
					"panic_type", panicType(recovered),
				)
			}
			logger.Info(
				"proxy-gateway http request",
				"request_id", requestID,
				"method", ctx.Request.Method,
				"path", ctx.Request.URL.Path,
				"status", ctx.Writer.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}()
		if opts.Authorize != nil && !opts.Authorize(ctx) {
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func writePanicError(w http.ResponseWriter, writeError PanicErrorWriter) {
	if writeError != nil {
		writeError(w)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func panicType(recovered any) string {
	if recovered == nil {
		return ""
	}
	return reflect.TypeOf(recovered).String()
}
