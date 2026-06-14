package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrCurrentLeaseActionRequired = errors.New("current lease action is required")

type CurrentLeaseAction func(context.Context, *proxyruntimev1.ProxyDynamicLease) error

type CurrentLeaseActionInput struct {
	Store      OrchestrationStore
	Locks      LockManager
	Lease      *proxyruntimev1.ProxyDynamicLease
	IsNotFound StoreNotFoundFunc
	Action     CurrentLeaseAction
}

func RunCurrentLeaseAction(ctx context.Context, input CurrentLeaseActionInput) error {
	if !HasLeaseID(input.Lease) {
		return nil
	}
	if input.Action == nil {
		return ErrCurrentLeaseActionRequired
	}
	return RunAccountAction(ctx, input.Locks, input.Lease.GetAccountId(), func(ctx context.Context) error {
		current, err := input.Store.LeaseFactByID(ctx, input.Lease.GetLeaseId())
		if err != nil {
			if storeNotFound(input.IsNotFound, err) {
				return nil
			}
			return err
		}
		return input.Action(ctx, current)
	})
}
