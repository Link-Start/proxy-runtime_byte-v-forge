package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (api *runtimeHTTPAPI) handleMetrics(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if api.metrics == nil {
		ctx.Status(http.StatusOK)
		return
	}
	if err := api.metrics.WritePrometheus(ctx.Writer); err != nil {
		writeHTTPError(ctx.Writer, appcore.InternalError("write metrics", err), http.StatusInternalServerError)
		return
	}
}

func (api *runtimeHTTPAPI) handleRuntimeMetricsSummary(ctx *gin.Context) {
	response, err := api.metricsUI.GetProxyRuntimeMetricsSummary(ctx.Request.Context())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}
