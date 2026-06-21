package app

import (
	"context"
	"io"
	"net/http"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/gin-gonic/gin"
)

type runtimeMetricsApplication struct {
	metrics *runtimeMetrics
	now     func() time.Time
}

type runtimeMetricsApplicationDependencies struct {
	Metrics *runtimeMetrics
	Now     func() time.Time
}

func newRuntimeMetricsApplication(deps runtimeMetricsApplicationDependencies) runtimeMetricsApplication {
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	return runtimeMetricsApplication{metrics: deps.Metrics, now: now}
}

func (a runtimeMetricsApplication) GetProxyGatewayMetricsSummary(context.Context) (*proxygatewayv1.GetProxyGatewayMetricsSummaryResponse, error) {
	return &proxygatewayv1.GetProxyGatewayMetricsSummaryResponse{
		Summary: a.metrics.Summary(a.now()),
	}, nil
}

func (a runtimeMetricsApplication) WritePrometheus(w io.Writer) error {
	if a.metrics == nil {
		return nil
	}
	return a.metrics.WritePrometheus(w)
}

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
	response, err := api.metricsUI.GetProxyGatewayMetricsSummary(ctx.Request.Context())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}
