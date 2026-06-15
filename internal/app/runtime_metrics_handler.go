package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleMetrics(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if api.service == nil || api.service.metrics == nil {
		ctx.Status(http.StatusOK)
		return
	}
	if err := api.service.metrics.WritePrometheus(ctx.Writer); err != nil {
		writeHTTPError(ctx.Writer, internalError("write metrics", err), http.StatusInternalServerError)
		return
	}
}
