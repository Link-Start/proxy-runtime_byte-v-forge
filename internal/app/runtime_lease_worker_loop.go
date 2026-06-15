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
	r.markLeaseWorkerStarted()
	err := errors.Join(
		r.runLeaseWorkerTask(ctx, "expire proxy leases", leaseCleanupAttemptTimeout, r.service().leases.ExpireDueLeaseFacts),
		r.runLeaseWorkerTask(ctx, "cleanup pending proxy leases", leaseCleanupAttemptTimeout, r.service().leases.CleanupPendingLeaseFacts),
	)
	r.markLeaseWorkerFinished(err)
}

func (r *Runtime) runLeaseWorkerTask(ctx context.Context, name string, timeout time.Duration, task leaseWorkerTask) error {
	if task == nil {
		return nil
	}
	startedAt := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := task(taskCtx)
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	r.logger.Warn(name+" failed", "error_type", errorLogType(err), "duration_ms", time.Since(startedAt).Milliseconds())
	return err
}
