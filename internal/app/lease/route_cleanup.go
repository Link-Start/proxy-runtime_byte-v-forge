package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
