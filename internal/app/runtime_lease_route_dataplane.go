package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func acquiredLeaseSessionRoute(session *proxyruntimev1.ProxySession, listener leaseapp.Listener, nodes []provider.Node, dialerProxy string, fallbackProtocol string) leaseapp.SessionRoute {
	return leaseapp.SessionRoute{
		SessionID:   session.GetSessionId(),
		Listener:    localServiceFromLeaseListener(listener, fallbackProtocol),
		Pool:        nodes,
		DialerProxy: dialerProxy,
	}
}

func (c leaseCoordinator) applyAcquiredLeaseDataPlaneRoute(ctx context.Context, route leaseapp.SessionRoute, failure *leaseAcquireFailure) error {
	if err := c.deps.dataPlane.UpsertSessionRoute(ctx, route); err != nil {
		failure.afterRoute(route, "dataplane route apply failed")
		return unavailable("dataplane route apply failed", err)
	}
	return nil
}
