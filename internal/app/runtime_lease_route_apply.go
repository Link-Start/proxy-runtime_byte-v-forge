package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) applyAcquiredLeaseRoute(ctx context.Context, flow acquiredLeaseFlow) (*proxyruntimev1.ProxyDynamicLease, error) {
	input := acquiredLeaseEndpointInput{
		advertisedHost:    flow.advertisedHost,
		req:               flow.request,
		settings:          flow.settings,
		selection:         flow.selection,
		providerAccountID: flow.providerAccountID,
		leaseID:           flow.leaseID,
		concurrencyHolder: flow.concurrencyHolder,
		providerClient:    flow.providerClient,
		session:           flow.session,
		lineLabels:        flow.lineLabels,
	}
	listener, listenerProto, egress, err := c.acquiredLeaseEndpoint(ctx, input, flow.failure)
	if err != nil {
		return nil, err
	}
	route := leaseapp.NewAcquiredSessionRoute(leaseapp.AcquiredSessionRouteInput{
		Session:       flow.session,
		Egress:        egress,
		Listener:      listener,
		Nodes:         flow.nodes,
		DialerProxy:   flow.dialerProxy,
		LocalProtocol: c.deps.cfg.LocalProtocol,
	})
	if err := c.applyAcquiredLeaseDataPlaneRoute(ctx, route, flow.failure); err != nil {
		return nil, err
	}
	lease, err := leaseapp.SaveAcquiredActiveFact(ctx, c.deps.store, leaseapp.AcquiredActiveFactInput{
		LeaseID:           flow.leaseID,
		Request:           flow.request,
		ProviderAccountID: flow.providerAccountID,
		Session:           flow.session,
		Egress:            egress,
		Listener:          listenerProto,
		SelectionPlan:     flow.selection.plan,
		AcquiredAt:        c.now(),
	})
	if err != nil {
		flow.failure.AfterRoute(ctx, route, "lease fact save failed")
		return nil, internalError("lease fact save failed", err)
	}
	c.clearExitCheckCache()
	if flow.request.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	return lease, nil
}
