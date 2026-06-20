package app

import (
	"context"
	"io"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func (a runtimeMetricsApplication) GetProxyRuntimeMetricsSummary(context.Context) (*proxyruntimev1.GetProxyRuntimeMetricsSummaryResponse, error) {
	return &proxyruntimev1.GetProxyRuntimeMetricsSummaryResponse{
		Summary: a.metrics.Summary(a.now()),
	}, nil
}

func (a runtimeMetricsApplication) WritePrometheus(w io.Writer) error {
	if a.metrics == nil {
		return nil
	}
	return a.metrics.WritePrometheus(w)
}
