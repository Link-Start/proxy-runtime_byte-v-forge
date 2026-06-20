package app

import (
	"context"
	"errors"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const (
	leaseExpirySweepInterval   = 30 * time.Second
	leaseCleanupAttemptTimeout = 20 * time.Second
)

type leaseWorkerTask func(context.Context) error

func (r *Runtime) leaseExpiryLoop(ctx context.Context) {
	r.runLeaseExpirySweep(ctx)
	ticker := time.NewTicker(leaseExpirySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runLeaseExpirySweep(ctx)
		}
	}
}

func (r *Runtime) runLeaseExpirySweep(ctx context.Context) {
	r.markLeaseWorkerStarted()
	err := errors.Join(
		r.runLeaseWorkerTask(ctx, runtimeMetricLeaseWorkerExpireDue, "expire proxy leases", leaseCleanupAttemptTimeout, r.service().leases.ExpireDueLeaseFacts),
		r.runLeaseWorkerTask(ctx, runtimeMetricLeaseWorkerCleanupPending, "cleanup pending proxy leases", leaseCleanupAttemptTimeout, r.service().leases.CleanupPendingLeaseFacts),
	)
	r.markLeaseWorkerFinished(err)
}

func (r *Runtime) runLeaseWorkerTask(ctx context.Context, operation string, name string, timeout time.Duration, task leaseWorkerTask) error {
	if task == nil {
		return nil
	}
	startedAt := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := task(taskCtx)
	r.observeRuntimeOperation(operation, startedAt, err)
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	r.logger.Warn(name+" failed", "error_type", appcore.ErrorLogType(err), "duration_ms", time.Since(startedAt).Milliseconds())
	return err
}
