package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) applyAcquiredLeaseRoute(ctx context.Context, flow acquiredLeaseFlow) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := leaseapp.ApplyAcquiredRouteFlow(ctx, leaseapp.AcquiredRouteFlowInput{
		Store:             c.deps.store,
		DataPlane:         c.deps.dataPlane,
		Failure:           flow.failure,
		LeaseID:           flow.leaseID,
		Request:           flow.request,
		ProviderClient:    flow.providerClient,
		ProviderAccountID: flow.providerAccountID,
		ConcurrencyHolder: flow.concurrencyHolder,
		Session:           flow.session,
		Nodes:             flow.nodes,
		DialerProxy:       flow.dialerProxy,
		LineLabels:        flow.lineLabels,
		LocalProtocol:     c.deps.cfg.LocalProtocol,
		SelectionPlan:     flow.selection.plan,
		AcquiredAt:        c.now(),
		Managed:           true,
		FallbackProtocol:  "http",
		ResolveListener: func(ctx context.Context, accountID string, leaseID string) (leaseapp.Listener, error) {
			return c.deps.leaseListener(ctx, flow.settings, accountID, leaseID)
		},
		ResolveEgress: func(ctx context.Context, listener leaseapp.Listener) (*proxyruntimev1.ProxyEndpoint, error) {
			_ = ctx
			return c.deps.localListenerEndpoint(listener, c.deps.sessionAdvertisedHost(flow.advertisedHost, listener))
		},
		AfterApply: func(ctx context.Context, _ *proxyruntimev1.ProxyDynamicLease) {
			c.clearExitCheckCache()
			if flow.request.GetAccountId() == playgroundProfileID {
				c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
			}
		},
	})
	if err != nil {
		return nil, acquiredRouteApplyError(err)
	}
	return lease, nil
}
