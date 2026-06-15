package app

import (
	"context"
	"errors"
	"time"
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
	r.runLeaseWorkerTask(ctx, "expire proxy leases", leaseCleanupAttemptTimeout, r.service().leases.ExpireDueLeaseFacts)
	r.runLeaseWorkerTask(ctx, "cleanup pending proxy leases", leaseCleanupAttemptTimeout, r.service().leases.CleanupPendingLeaseFacts)
}

func (r *Runtime) runLeaseWorkerTask(ctx context.Context, name string, timeout time.Duration, task leaseWorkerTask) {
	if task == nil {
		return
	}
	startedAt := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := task(taskCtx); err != nil && !errors.Is(err, context.Canceled) {
		r.logger.Warn(name+" failed", "error", err, "duration_ms", time.Since(startedAt).Milliseconds())
	}
}
