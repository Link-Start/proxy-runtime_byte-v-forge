package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (c leaseCoordinator) Acquire(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	runner := c.preparedAcquireRunner(advertisedHost, req)
	lease, err := runner.Run(ctx, leaseapp.PreparedAcquireRunnerInput{
		Request: req,
	})
	if err != nil && leaseapp.IsAcquireRequestError(err) {
		return nil, appcore.InvalidArgument(err.Error(), err)
	}
	return lease, err
}

func (c leaseCoordinator) Release(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := c.releaseRunner().Release(ctx, req)
	if err != nil && leaseapp.IsReleaseLookupRequestError(err) {
		return nil, appcore.InvalidArgument(err.Error(), err)
	}
	return lease, err
}
