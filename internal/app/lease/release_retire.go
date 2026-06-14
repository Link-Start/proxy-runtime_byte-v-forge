package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrReleaseRetireActionRequired = errors.New("release retire action is required")

type ReleaseRetireAction func(context.Context, *proxyruntimev1.ProxyDynamicLease) error

type ReleaseRetireInput struct {
	Store      OrchestrationStore
	Locks      LockManager
	Lease      *proxyruntimev1.ProxyDynamicLease
	IsNotFound StoreNotFoundFunc
	Retire     ReleaseRetireAction
}

func RetireReleaseLease(ctx context.Context, input ReleaseRetireInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease := input.Lease
	if !ReleaseNeedsRouteRetire(lease) {
		return lease, nil
	}
	if input.Retire == nil {
		return lease, ErrReleaseRetireActionRequired
	}
	err := WithAccountLock(ctx, input.Locks, lease.GetAccountId(), func(ctx context.Context) error {
		current, err := RefreshReleaseLease(ctx, input.Store, lease, input.IsNotFound)
		lease = current
		if err != nil || !ReleaseNeedsRouteRetire(lease) {
			return err
		}
		return input.Retire(ctx, lease)
	})
	return lease, err
}
