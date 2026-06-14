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
	lease, err := leaseapp.ApplyAcquiredEndpointRoute(ctx, leaseapp.AcquiredEndpointRouteApplyInput{
		Store:             c.deps.store,
		DataPlane:         c.deps.dataPlane,
		Failure:           flow.failure,
		LeaseID:           flow.leaseID,
		Request:           flow.request,
		ProviderAccountID: flow.providerAccountID,
		Session:           flow.session,
		Egress:            egress,
		Listener:          listener,
		ListenerProto:     listenerProto,
		Nodes:             flow.nodes,
		DialerProxy:       flow.dialerProxy,
		LocalProtocol:     c.deps.cfg.LocalProtocol,
		SelectionPlan:     flow.selection.plan,
		AcquiredAt:        c.now(),
	})
	if err != nil {
		return nil, acquiredRouteApplyError(err)
	}
	c.clearExitCheckCache()
	if flow.request.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	return lease, nil
}
