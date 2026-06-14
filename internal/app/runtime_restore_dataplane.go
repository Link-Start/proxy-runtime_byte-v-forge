package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (c leaseCoordinator) restoreLeaseDataPlaneRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, settings *runtimeSettingsFile, nodes []provider.Node) error {
	dialerProxy, lineLabels, err := c.deps.dynamicLeaseDialerProxy(ctx, settings, lease.GetAccountId())
	if err != nil {
		return err
	}
	route, ok := leaseapp.SessionRouteFromLease(lease, leaseapp.ApplyNodeLabels(nodes, lineLabels), dialerProxy, c.deps.cfg.LocalProtocol)
	if !ok {
		return nil
	}
	return leaseapp.UpsertSessionRoute(ctx, c.deps.dataPlane, route)
}
