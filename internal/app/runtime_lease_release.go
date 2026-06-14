package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) releaseLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	retirer := c.leaseRouteRetirer()
	lease, err := leaseapp.ReleaseLease(ctx, leaseapp.ReleaseInput{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		Request:    req,
		IsNotFound: isStoreNotFound,
		Retire:     retirer.Retire,
	})
	if err != nil && leaseapp.IsReleaseLookupRequestError(err) {
		return nil, invalidArgument(err.Error(), err)
	}
	return lease, err
}
