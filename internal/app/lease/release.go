package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ReleaseInput struct {
	Store      OrchestrationStore
	Locks      LockManager
	Request    *proxyruntimev1.ReleaseProxyLeaseRequest
	IsNotFound StoreNotFoundFunc
	Retire     ReleaseRetireAction
}

func ReleaseLease(ctx context.Context, input ReleaseInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := LookupReleaseLease(ctx, input.Store, input.Request, input.IsNotFound)
	if err != nil {
		return nil, err
	}
	return RetireReleaseLease(ctx, ReleaseRetireInput{
		Store:      input.Store,
		Locks:      input.Locks,
		Lease:      lease,
		IsNotFound: input.IsNotFound,
		Retire:     input.Retire,
	})
}
