package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) leaseByReleaseRequest(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := leaseapp.LookupReleaseLease(ctx, c.deps.store, req, isStoreNotFound)
	if err == nil {
		return lease, nil
	}
	if leaseapp.IsReleaseLookupRequestError(err) {
		return nil, invalidArgument(err.Error(), err)
	}
	return nil, err
}
