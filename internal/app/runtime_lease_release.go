package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) releaseLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := c.leaseByReleaseRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return leaseapp.RetireReleaseLease(ctx, leaseapp.ReleaseRetireInput{
		Store:      c.deps.store,
		Locks:      c.deps.locks,
		Lease:      lease,
		IsNotFound: isStoreNotFound,
		Retire:     c.retireLeaseRoute,
	})
}
