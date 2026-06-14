package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrSessionListenerAllocationActionRequired = errors.New("session listener allocation action is required")

type SessionListenerAllocationAction func(context.Context) (*proxyruntimev1.ProxyDynamicLease, error)

func RunSessionListenerAllocation(ctx context.Context, locks LockManager, action SessionListenerAllocationAction) (*proxyruntimev1.ProxyDynamicLease, error) {
	if action == nil {
		return nil, ErrSessionListenerAllocationActionRequired
	}
	var lease *proxyruntimev1.ProxyDynamicLease
	err := WithSessionListenerAllocationLock(ctx, locks, func(ctx context.Context) error {
		var err error
		lease, err = action(ctx)
		return err
	})
	return lease, err
}
