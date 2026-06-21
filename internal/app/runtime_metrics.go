package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"sync"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	runtimeMetricLeaseList                   = "lease_list"
	runtimeMetricLeaseAcquire                = "lease_acquire"
	runtimeMetricLeaseRelease                = "lease_release"
	runtimeMetricLeaseWorkerRestoreActive    = "lease_worker_restore_active"
	runtimeMetricLeaseWorkerExpireDue        = "lease_worker_expire_due"
	runtimeMetricLeaseWorkerCleanupPending   = "lease_worker_cleanup_pending"
	runtimeMetricProviderFetchBase           = "provider_fetch_base"
	runtimeMetricProviderSessionFactory      = "provider_session_factory"
	runtimeMetricProviderSessionCreate       = "provider_session_create"
	runtimeMetricProviderSessionFetch        = "provider_session_fetch"
	runtimeMetricProviderSessionRelease      = "provider_session_release"
	runtimeMetricDataPlaneApplyDesiredConfig = "dataplane_apply_desired_config"
	runtimeMetricDataPlaneUpsertSessionRoute = "dataplane_upsert_session_route"
	runtimeMetricDataPlaneDeleteSessionRoute = "dataplane_delete_session_route"
	runtimeMetricSettingsApply               = "settings_apply"
)

const runtimeMetricSlowThreshold = time.Second

type runtimeMetrics struct {
	mu            sync.Mutex
	operations    map[runtimeMetricKey]*runtimeMetricSample
	slowThreshold time.Duration
}

type runtimeMetricKey struct {
	operation string
	status    string
}

type runtimeMetricSample struct {
	count          uint64
	slowCount      uint64
	durationSecond float64
}

func newRuntimeMetrics() *runtimeMetrics {
	return &runtimeMetrics{
		operations:    map[runtimeMetricKey]*runtimeMetricSample{},
		slowThreshold: runtimeMetricSlowThreshold,
	}
}

func (m *runtimeMetrics) Observe(operation string, startedAt time.Time, err error) {
	if m == nil {
		return
	}
	duration := time.Since(startedAt)
	key := runtimeMetricKey{operation: runtimeMetricLabelValue(operation), status: runtimeMetricStatus(err)}
	m.mu.Lock()
	defer m.mu.Unlock()
	sample := m.operations[key]
	if sample == nil {
		sample = &runtimeMetricSample{}
		m.operations[key] = sample
	}
	sample.count++
	if duration >= m.slowThreshold {
		sample.slowCount++
	}
	sample.durationSecond += duration.Seconds()
}

func (m *runtimeMetrics) WritePrometheus(w io.Writer) error {
	if m == nil {
		return nil
	}
	samples := m.snapshot()
	if _, err := io.WriteString(w, "# HELP proxy_gateway_operation_total Total runtime operations by operation and status.\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "# TYPE proxy_gateway_operation_total counter\n"); err != nil {
		return err
	}
	for _, item := range samples {
		if _, err := fmt.Fprintf(w, "proxy_gateway_operation_total%s %d\n", item.labels(), item.sample.count); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "# HELP proxy_gateway_operation_duration_seconds Runtime operation duration summary by operation and status.\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "# TYPE proxy_gateway_operation_duration_seconds summary\n"); err != nil {
		return err
	}
	for _, item := range samples {
		if _, err := fmt.Fprintf(w, "proxy_gateway_operation_duration_seconds_sum%s %.6f\n", item.labels(), item.sample.durationSecond); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "proxy_gateway_operation_duration_seconds_count%s %d\n", item.labels(), item.sample.count); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "# HELP proxy_gateway_operation_slow_total Total runtime operations whose duration reached the fixed slow threshold.\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "# TYPE proxy_gateway_operation_slow_total counter\n"); err != nil {
		return err
	}
	for _, item := range samples {
		if _, err := fmt.Fprintf(w, "proxy_gateway_operation_slow_total%s %d\n", item.labels(), item.sample.slowCount); err != nil {
			return err
		}
	}
	return nil
}

func (m *runtimeMetrics) snapshot() []runtimeMetricSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]runtimeMetricSnapshot, 0, len(m.operations))
	for key, sample := range m.operations {
		out = append(out, runtimeMetricSnapshot{
			key: key,
			sample: runtimeMetricSample{
				count:          sample.count,
				slowCount:      sample.slowCount,
				durationSecond: sample.durationSecond,
			},
		})
	}
	sort.Slice(out, func(i int, j int) bool {
		if out[i].key.operation == out[j].key.operation {
			return out[i].key.status < out[j].key.status
		}
		return out[i].key.operation < out[j].key.operation
	})
	return out
}

type runtimeMetricSnapshot struct {
	key    runtimeMetricKey
	sample runtimeMetricSample
}

func (s runtimeMetricSnapshot) labels() string {
	return "{operation=" + strconv.Quote(s.key.operation) + ",status=" + strconv.Quote(s.key.status) + "}"
}

func runtimeMetricStatus(err error) string {
	switch {
	case err == nil:
		return "success"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		return "error"
	}
}

func runtimeMetricLabelValue(value string) string {
	if value == "" {
		return "unknown"
	}
	out := make([]rune, 0, len(value))
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '_' || r == '-':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

func (m *runtimeMetrics) Summary(now time.Time) *proxygatewayv1.ProxyGatewayMetricsSummary {
	summary := &proxygatewayv1.ProxyGatewayMetricsSummary{
		UpdatedAt: timestamppb.New(now),
	}
	if m == nil {
		return summary
	}
	summary.SlowThresholdSeconds = m.slowThreshold.Seconds()
	snapshots := m.snapshot()
	summary.Operations = make([]*proxygatewayv1.ProxyGatewayOperationMetric, 0, len(snapshots))
	for _, snapshot := range snapshots {
		summary.Operations = append(summary.Operations, &proxygatewayv1.ProxyGatewayOperationMetric{
			Operation:       snapshot.key.operation,
			Status:          snapshot.key.status,
			Count:           snapshot.sample.count,
			SlowCount:       snapshot.sample.slowCount,
			DurationSeconds: snapshot.sample.durationSecond,
		})
	}
	return summary
}
