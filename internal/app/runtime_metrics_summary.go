package app

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *runtimeMetrics) Summary(now time.Time) *proxyruntimev1.ProxyRuntimeMetricsSummary {
	summary := &proxyruntimev1.ProxyRuntimeMetricsSummary{
		UpdatedAt: timestamppb.New(now),
	}
	if m == nil {
		return summary
	}
	summary.SlowThresholdSeconds = m.slowThreshold.Seconds()
	snapshots := m.snapshot()
	summary.Operations = make([]*proxyruntimev1.ProxyRuntimeOperationMetric, 0, len(snapshots))
	for _, snapshot := range snapshots {
		summary.Operations = append(summary.Operations, &proxyruntimev1.ProxyRuntimeOperationMetric{
			Operation:       snapshot.key.operation,
			Status:          snapshot.key.status,
			Count:           snapshot.sample.count,
			SlowCount:       snapshot.sample.slowCount,
			DurationSeconds: snapshot.sample.durationSecond,
		})
	}
	return summary
}
