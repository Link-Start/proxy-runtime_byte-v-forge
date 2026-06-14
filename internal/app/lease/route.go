package lease

import (
	"context"
	"errors"
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
