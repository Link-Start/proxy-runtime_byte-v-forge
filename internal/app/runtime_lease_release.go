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
	if !leaseapp.ReleaseNeedsRouteRetire(lease) {
		return lease, nil
	}
	accountID := strings.TrimSpace(lease.GetAccountId())
	err = leaseapp.WithAccountLock(ctx, c.deps.locks, accountID, func(ctx context.Context) error {
		lease, err = leaseapp.RefreshReleaseLease(ctx, c.deps.store, lease, isStoreNotFound)
		if err != nil || !leaseapp.ReleaseNeedsRouteRetire(lease) {
			return err
		}
		return c.retireLeaseRoute(ctx, lease)
	})
	return lease, err
}
