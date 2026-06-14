package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) leaseByReleaseRequest(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lookup, err := leaseapp.ParseReleaseRequest(req)
	if err != nil {
		return nil, invalidArgument(err.Error(), err)
	}
	if lookup.LeaseID != "" {
		lease, err := c.deps.store.LeaseFactByID(ctx, lookup.LeaseID)
		if err != nil {
			if isStoreNotFound(err) {
				return nil, invalidArgument("lease_id not found", nil)
			}
			return nil, err
		}
		if err := leaseapp.ValidateReleaseLeaseMatch(lookup, lease); err != nil {
			return nil, invalidArgument(err.Error(), err)
		}
		return lease, nil
	}
	lease, err := c.deps.store.ActiveLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err == nil {
		return lease, nil
	}
	if !isStoreNotFound(err) {
		return nil, err
	}
	lease, err = c.deps.store.LatestLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err != nil {
		if isStoreNotFound(err) {
			return nil, invalidArgument("active lease not found", nil)
		}
		return nil, err
	}
	return lease, nil
}
