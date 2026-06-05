package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

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
		if !leaseActive(lease, now) {
			if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
				if err := c.deleteLeaseRoute(ctx, lease); err != nil {
					r.logger.Warn("delete expired proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
					restoreErrors = append(restoreErrors, fmt.Errorf("delete expired lease route %q: %w", lease.GetLeaseId(), err))
				}
				lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED
				if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
					restoreErrors = append(restoreErrors, fmt.Errorf("save expired lease fact %q: %w", lease.GetLeaseId(), err))
				}
			}
			continue
		}
		if err := c.restoreLeaseRoute(ctx, lease); err != nil {
			r.logger.Warn("restore proxy lease route failed", "account_id", lease.GetAccountId(), "error", err)
			restoreErrors = append(restoreErrors, fmt.Errorf("restore lease route %q: %w", lease.GetLeaseId(), err))
			continue
		}
	}
	if err := c.cleanupPendingLeaseFacts(ctx); err != nil {
		restoreErrors = append(restoreErrors, err)
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
	settings, err := r.settings.load(ctx)
	if err != nil {
		return err
	}
	providerCfg.Gateways = gatewaysForDynamicGateway(settings, lease.GetRoutePlan(), providerCfg.ProviderID)
	providerClient, err := r.accountProviders.NewSessionProvider(providerCfg, BuildProviderHTTPClient(r.cfg))
	if err != nil {
		return err
	}
	nodes, err := providerClient.FetchSession(ctx, lease.GetSession())
	if err != nil {
		return err
	}
	route := dataplane.SessionRoute{
		SessionID: lease.GetSession().GetSessionId(),
		Listener:  localServiceFromListener(listenerFromProto(lease.GetListener()), r.cfg.LocalProtocol),
		Pool:      nodes,
	}
	return r.dataPlane.UpsertSessionRoute(ctx, route)
}
