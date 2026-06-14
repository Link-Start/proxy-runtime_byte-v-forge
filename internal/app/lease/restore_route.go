package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type RestoreRouteInput struct {
	DataPlane          DataPlaneApplier
	Lease              *proxyruntimev1.ProxyDynamicLease
	Nodes              []provider.Node
	LocalProtocol      string
	ResolveLineBinding RouteLineBindingResolver
}

func RestoreLeaseRoute(ctx context.Context, input RestoreRouteInput) error {
	lineBinding, err := PrepareRouteLineBinding(ctx, input.Nodes, input.Lease.GetAccountId(), input.ResolveLineBinding)
	if err != nil {
		return err
	}
	route, ok := SessionRouteFromLease(input.Lease, lineBinding.Nodes, lineBinding.DialerProxy, input.LocalProtocol)
	if !ok {
		return nil
	}
	return UpsertSessionRoute(ctx, input.DataPlane, route)
}
