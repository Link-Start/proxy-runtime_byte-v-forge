package app

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
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
		r.logger.Warn("background proxy lease restore failed", "error", err, "duration_ms", time.Since(startedAt).Milliseconds())
		return
	}
	r.logger.Info("background proxy lease restore finished", "duration_ms", time.Since(startedAt).Milliseconds())
}

func (c leaseCoordinator) restoreActiveLeases(ctx context.Context) error {
	return leaseapp.ProcessRestorableActiveFacts(ctx, leaseapp.WorkerBatchInput{
		Store:   c.deps.store,
		Timeout: leaseRestoreRouteTimeout,
		Process: c.restoreLeaseRoute,
		Observe: func(lease *proxyruntimev1.ProxyDynamicLease, err error) {
			c.warn("restore proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
		},
		ObserveList: func(err error) {
			c.warn("list proxy leases for restore failed", "error", err)
		},
	}, c.now().UTC())
}

func (c leaseCoordinator) restoreLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease.GetSession() == nil || lease.GetListener() == nil {
		return errors.New("lease session or listener is missing")
	}
	inputs, err := c.loadRestoreLeaseInputs(ctx, lease)
	if err != nil {
		return err
	}
	slot, err := c.acquireRestoreLeaseConcurrencySlot(ctx, lease, inputs.settings, inputs.providerAccount)
	if err != nil {
		return err
	}
	return leaseapp.RunTemporaryConcurrencySlot(ctx, leaseapp.TemporaryConcurrencySlotInput{
		Slot:           slot,
		ReleaseTimeout: leaseRestoreSlotReleaseTimeout,
		Action: func(ctx context.Context) error {
			nodes, err := c.restoreLeaseSessionNodes(ctx, lease, inputs.settings, inputs.providerConfig)
			if err != nil {
				return err
			}
			return c.restoreLeaseDataPlaneRoute(ctx, lease, inputs.settings, nodes)
		},
	})
}
