package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (c leaseCoordinator) applyAcquiredLeaseRoute(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest, settings *runtimeSettingsFile, selection dynamicIPSelection, providerAccountID string, leaseID string, concurrencyHolder string, providerClient leaseapp.SessionProvider, session *proxyruntimev1.ProxySession, nodes []provider.Node, dialerProxy string, lineLabels map[string]string, failure *leaseAcquireFailure) (*proxyruntimev1.ProxyDynamicLease, error) {
	input := acquiredLeaseEndpointInput{
		advertisedHost:    advertisedHost,
		req:               req,
		settings:          settings,
		selection:         selection,
		providerAccountID: providerAccountID,
		leaseID:           leaseID,
		concurrencyHolder: concurrencyHolder,
		providerClient:    providerClient,
		session:           session,
		lineLabels:        lineLabels,
	}
	listener, listenerProto, egress, err := c.acquiredLeaseEndpoint(ctx, input, failure)
	if err != nil {
		return nil, err
	}
	session.Egress = egress
	route := acquiredLeaseSessionRoute(session, listener, nodes, dialerProxy, c.deps.cfg.LocalProtocol)
	if err := c.applyAcquiredLeaseDataPlaneRoute(ctx, route, failure); err != nil {
		return nil, err
	}
	lease := leaseapp.NewActiveFact(leaseapp.ActiveFactInput{
		LeaseID:           leaseID,
		AccountID:         req.GetAccountId(),
		Purpose:           req.GetPurpose(),
		ProviderAccountID: providerAccountID,
		Session:           session,
		Egress:            egress,
		Listener:          listenerProto,
		SelectionPlan:     selection.plan,
		AcquiredAt:        c.now(),
	})
	if err := c.deps.store.SaveLeaseFact(ctx, lease); err != nil {
		failure.afterRoute(route, "lease fact save failed")
		return nil, internalError("lease fact save failed", err)
	}
	c.clearExitCheckCache()
	if req.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	return lease, nil
}
