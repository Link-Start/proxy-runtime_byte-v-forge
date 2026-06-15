package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleMetrics(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if api.service == nil {
		ctx.Status(http.StatusOK)
		return
	}
	if err := api.service.metrics.WritePrometheus(ctx.Writer); err != nil {
		writeHTTPError(ctx.Writer, internalError("write metrics", err), http.StatusInternalServerError)
		return
	}
}

func (api *runtimeHTTPAPI) handleRuntimeMetricsSummary(ctx *gin.Context) {
	response, err := api.service.GetProxyRuntimeMetricsSummary(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeMetricsSummaryRequest{})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}
