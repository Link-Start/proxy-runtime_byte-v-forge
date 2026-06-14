package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAccountLeaseActionRequired = errors.New("account lease action is required")

type AccountLeaseAction func(context.Context) (*proxyruntimev1.ProxyDynamicLease, error)

func RunAccountLeaseAction(ctx context.Context, locks LockManager, accountID string, action AccountLeaseAction) (*proxyruntimev1.ProxyDynamicLease, error) {
	if action == nil {
		return nil, ErrAccountLeaseActionRequired
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err := WithAccountLock(ctx, locks, accountID, func(ctx context.Context) error {
		var err error
		lease, err = action(ctx)
		return err
	})
	return lease, err
}
