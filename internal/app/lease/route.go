package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

var ErrDataPlaneApplierRequired = errors.New("lease dataplane applier is required")

func UpsertSessionRoute(ctx context.Context, dataPlane DataPlaneApplier, route SessionRoute) error {
	if dataPlane == nil {
		return ErrDataPlaneApplierRequired
	}
	return dataPlane.UpsertSessionRoute(ctx, route)
}

func DeleteSessionRoute(ctx context.Context, dataPlane DataPlaneApplier, route SessionRoute) error {
	if dataPlane == nil {
		return ErrDataPlaneApplierRequired
	}
	return dataPlane.DeleteSessionRoute(ctx, route)
}

func DeleteLeaseRoute(ctx context.Context, dataPlane DataPlaneApplier, lease *proxyruntimev1.ProxyDynamicLease, localProtocol string) error {
	route, ok := SessionRouteFromLease(lease, nil, "", localProtocol)
	if !ok {
		return nil
	}
	return DeleteSessionRoute(ctx, dataPlane, route)
}

type RouteCleanupFailureRecorder func(context.Context, *proxyruntimev1.ProxyDynamicLease) error

type RouteCleanupInput struct {
	DataPlane     DataPlaneApplier
	Lease         *proxyruntimev1.ProxyDynamicLease
	LocalProtocol string
	RecordFailure RouteCleanupFailureRecorder
}

func CleanupLeaseRoute(ctx context.Context, input RouteCleanupInput) error {
	if err := DeleteLeaseRoute(ctx, input.DataPlane, input.Lease, input.LocalProtocol); err != nil {
		if input.RecordFailure != nil {
			_ = input.RecordFailure(ctx, input.Lease)
		}
		return err
	}
	return nil
}

var ErrRouteLineBindingResolverRequired = errors.New("route line binding resolver is required")

type RouteLineBindingResolver func(context.Context, string) (string, map[string]string, error)

type RouteLineBinding struct {
	Nodes       []provider.Node
	DialerProxy string
	Labels      map[string]string
}

func PrepareRouteLineBinding(ctx context.Context, nodes []provider.Node, accountID string, resolve RouteLineBindingResolver) (RouteLineBinding, error) {
	if resolve == nil {
		return RouteLineBinding{}, ErrRouteLineBindingResolverRequired
	}
	dialerProxy, labels, err := resolve(ctx, accountID)
	if err != nil {
		return RouteLineBinding{}, err
	}
	return RouteLineBinding{Nodes: ApplyNodeLabels(nodes, labels), DialerProxy: dialerProxy, Labels: labels}, nil
}
