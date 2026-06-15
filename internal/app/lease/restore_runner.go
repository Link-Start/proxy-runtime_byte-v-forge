package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrLeaseRouteRestorerRequired = errors.New("lease route restorer is required")

type LeaseRouteRestorerResolver func(context.Context, *proxyruntimev1.ProxyDynamicLease) (LeaseRouteRestorer, error)

type RestoreLeaseRouteRunner struct {
	ResolveRestorer LeaseRouteRestorerResolver
}

func (r RestoreLeaseRouteRunner) Restore(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if r.ResolveRestorer == nil {
		return ErrLeaseRouteRestorerRequired
	}
	restorer, err := r.ResolveRestorer(ctx, lease)
	if err != nil {
		return err
	}
	return restorer.Restore(ctx, lease)
}
