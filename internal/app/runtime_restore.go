package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
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
	if c.deps.store == nil {
		return nil
	}
	leases, err := c.deps.store.ListRestorableLeaseFacts(ctx)
	if err != nil {
		c.warn("list proxy leases for restore failed", "error", err)
		return err
	}
	now := c.now().UTC()
	restoreErrors := make([]error, 0)
	for _, lease := range leases {
		if !leaseapp.ActiveAt(lease, now) {
			continue
		}
		if err := ctx.Err(); err != nil {
			restoreErrors = append(restoreErrors, err)
			break
		}
		leaseCtx, cancel := context.WithTimeout(ctx, leaseRestoreRouteTimeout)
		err := c.restoreLeaseRoute(leaseCtx, lease)
		cancel()
		if err != nil {
			c.warn("restore proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
			restoreErrors = append(restoreErrors, fmt.Errorf("restore lease route %q: %w", lease.GetLeaseId(), err))
			continue
		}
	}
	return errors.Join(restoreErrors...)
}

func (c leaseCoordinator) restoreLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease.GetSession() == nil || lease.GetListener() == nil {
		return errors.New("lease session or listener is missing")
	}
	providerCfg, _, err := c.deps.store.ProviderConfig(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	providerAccount, err := c.deps.store.ProviderAccount(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return err
	}
	holder := leaseConcurrencyHolder(lease)
	policy := leaseConcurrencyPolicy(lease)
	slot, err := c.acquireProviderAccountConcurrencySlot(ctx, providerAccount, dynamicProviderConcurrencyLimit(settings, leaseDynamicProviderID(lease), policy), policy, holder, leaseConcurrencySlotTTL(policy))
	if err != nil {
		return err
	}
	keepSlot := false
	defer func() {
		if !keepSlot {
			releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), leaseRestoreSlotReleaseTimeout)
			defer cancel()
			_ = slot.Release(releaseCtx)
		}
	}()
	providerCfg.Gateways = endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerCfg.ProviderID)
	providerClient, err := c.newSessionProvider(providerCfg)
	if err != nil {
		return err
	}
	nodes, err := providerClient.FetchSession(ctx, lease.GetSession())
	if err != nil {
		return err
	}
	dialerProxy, lineLabels, err := c.deps.dynamicLeaseDialerProxy(ctx, settings, lease.GetAccountId())
	if err != nil {
		return err
	}
	nodes = applyDynamicLeaseLineLabels(nodes, lineLabels)
	route := dataplane.SessionRoute{
		SessionID:   lease.GetSession().GetSessionId(),
		Listener:    localServiceFromListener(listenerFromProto(lease.GetListener()), c.deps.cfg.LocalProtocol),
		Pool:        nodes,
		DialerProxy: dialerProxy,
	}
	if err := c.deps.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		return err
	}
	keepSlot = true
	return nil
}
