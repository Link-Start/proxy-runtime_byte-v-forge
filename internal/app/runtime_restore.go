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

const startupLeaseRestoreTimeout = 2 * time.Minute

func (r *Runtime) restoreActiveLeasesInBackground(ctx context.Context) {
	r.markLeaseRestoreStarted()
	startedAt := time.Now()
	restoreCtx, cancel := context.WithTimeout(ctx, startupLeaseRestoreTimeout)
	defer cancel()
	err := r.leaseCoordinator.restoreActiveLeases(restoreCtx)
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
	r := c.runtime
	if r.store == nil {
		return nil
	}
	leases, err := r.store.ListRestorableLeaseFacts(ctx)
	if err != nil {
		r.logger.Warn("list proxy leases for restore failed", "error", err)
		return err
	}
	now := time.Now().UTC()
	restoreErrors := make([]error, 0)
	for _, lease := range leases {
		if !leaseapp.ActiveAt(lease, now) {
			continue
		}
		if err := c.restoreLeaseRoute(ctx, lease); err != nil {
			r.logger.Warn("restore proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
			restoreErrors = append(restoreErrors, fmt.Errorf("restore lease route %q: %w", lease.GetLeaseId(), err))
			continue
		}
	}
	return errors.Join(restoreErrors...)
}

func (c leaseCoordinator) restoreLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	if lease.GetSession() == nil || lease.GetListener() == nil {
		return errors.New("lease session or listener is missing")
	}
	providerCfg, _, err := r.store.ProviderConfig(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	providerAccount, err := r.store.ProviderAccount(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		return err
	}
	holder := leaseConcurrencyHolder(lease)
	policy := leaseConcurrencyPolicy(lease)
	slot, err := r.acquireProviderAccountConcurrencySlot(ctx, providerAccount, dynamicProviderConcurrencyLimit(settings, leaseDynamicProviderID(lease), policy), policy, holder, leaseConcurrencySlotTTL(policy))
	if err != nil {
		return err
	}
	keepSlot := false
	defer func() {
		if !keepSlot {
			_ = slot.Release(ctx)
		}
	}()
	providerCfg.Gateways = endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerCfg.ProviderID)
	providerClient, err := r.accountProviders.NewSessionProvider(providerCfg, r.providerHTTPClient)
	if err != nil {
		return err
	}
	nodes, err := providerClient.FetchSession(ctx, lease.GetSession())
	if err != nil {
		return err
	}
	dialerProxy, lineLabels, err := r.dynamicLeaseDialerProxy(ctx, settings, lease.GetAccountId())
	if err != nil {
		return err
	}
	nodes = applyDynamicLeaseLineLabels(nodes, lineLabels)
	route := dataplane.SessionRoute{
		SessionID:   lease.GetSession().GetSessionId(),
		Listener:    localServiceFromListener(listenerFromProto(lease.GetListener()), r.cfg.LocalProtocol),
		Pool:        nodes,
		DialerProxy: dialerProxy,
	}
	if err := r.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		return err
	}
	keepSlot = true
	return nil
}
