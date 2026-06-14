package app

import (
	"context"
	"time"
)

const (
	leaseExpirySweepInterval   = 30 * time.Second
	leaseCleanupAttemptTimeout = 20 * time.Second
)

func (r *Runtime) leaseExpiryLoop(ctx context.Context) {
	ticker := time.NewTicker(leaseExpirySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.service().leases.ExpireDueLeaseFacts(ctx); err != nil {
				r.logger.Warn("expire proxy leases failed", "error", err)
			}
			if err := r.service().leases.CleanupPendingLeaseFacts(ctx); err != nil {
				r.logger.Warn("cleanup pending proxy leases failed", "error", err)
			}
		}
	}
}
