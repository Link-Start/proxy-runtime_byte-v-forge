package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (c leaseCoordinator) restoreLeaseDataPlaneRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, settings *runtimeSettingsFile, nodes []provider.Node) error {
	return leaseapp.RestoreLeaseRoute(ctx, leaseapp.RestoreRouteInput{
		DataPlane:     c.deps.dataPlane,
		Lease:         lease,
		Nodes:         nodes,
		LocalProtocol: c.deps.cfg.LocalProtocol,
		ResolveLineBinding: func(ctx context.Context, accountID string) (string, map[string]string, error) {
			return c.deps.dynamicLeaseDialerProxy(ctx, settings, accountID)
		},
	})
}
