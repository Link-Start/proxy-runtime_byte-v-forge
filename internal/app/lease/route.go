package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
