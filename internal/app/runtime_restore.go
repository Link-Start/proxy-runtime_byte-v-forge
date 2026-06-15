package app

import (
	"context"
	"errors"
	"time"
)

const (
	startupLeaseRestoreTimeout     = 2 * time.Minute
	leaseRestoreRouteTimeout       = 20 * time.Second
	leaseRestoreSlotReleaseTimeout = 5 * time.Second
)

func (r *Runtime) restoreActiveLeasesInBackground(ctx context.Context) {
	r.markLeaseRestoreStarted()
	startedAt := time.Now()
	restoreCtx, cancel := context.WithTimeout(ctx, startupLeaseRestoreTimeout)
	defer cancel()
	err := r.service().leases.RestoreActiveLeases(restoreCtx)
	r.markLeaseRestoreFinished(err)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			r.logger.Info("background proxy lease restore stopped", "duration_ms", time.Since(startedAt).Milliseconds())
			return
		}
		r.logger.Warn("background proxy lease restore failed", "error_type", errorLogType(err), "duration_ms", time.Since(startedAt).Milliseconds())
		return
	}
	r.logger.Info("background proxy lease restore finished", "duration_ms", time.Since(startedAt).Milliseconds())
}
