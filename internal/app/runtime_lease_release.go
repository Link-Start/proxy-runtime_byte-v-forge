package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) releaseLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := c.leaseByReleaseRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if leaseapp.HasReleasedStatus(lease) {
		return lease, nil
	}
	accountID := strings.TrimSpace(lease.GetAccountId())
	err = leaseapp.WithAccountLock(ctx, c.deps.locks, accountID, func(ctx context.Context) error {
		current, err := c.deps.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil && !isStoreNotFound(err) {
			return err
		}
		if current != nil {
			lease = current
		}
		if leaseapp.HasReleasedStatus(lease) {
			return nil
		}
		if !leaseapp.HasActiveStatus(lease) {
			return nil
		}
		return c.retireLeaseRoute(ctx, lease)
	})
	return lease, err
}
