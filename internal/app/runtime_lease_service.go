package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireLease(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := leaseapp.RunPreparedAcquire(ctx, leaseapp.PreparedAcquireInput{
		Locks:   c.deps.locks,
		Request: req,
		Action: func(ctx context.Context) (*proxyruntimev1.ProxyDynamicLease, error) {
			return c.acquireLeaseWithAccountLock(ctx, advertisedHost, req)
		},
	})
	if err != nil && leaseapp.IsAcquireRequestError(err) {
		return nil, invalidArgument(err.Error(), err)
	}
	return lease, err
}

func (c leaseCoordinator) acquireLeaseWithAccountLock(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	runner := c.accountLockedAcquireRunner(ctx, settings, advertisedHost, req)
	lease, err := runner.Run(ctx, leaseapp.AccountLockedAcquireRunnerInput{
		Request:        req,
		EgressProfiles: settings.GetEgressProfiles(),
	})
	if err != nil && leaseapp.IsAcquirePolicyError(err) {
		return nil, leaseProfilePolicyError(err)
	}
	return lease, err
}
