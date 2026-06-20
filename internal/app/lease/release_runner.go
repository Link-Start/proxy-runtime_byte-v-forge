package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ReleaseRunner struct {
	Store      OrchestrationStore
	Locks      LockManager
	IsNotFound StoreNotFoundFunc
	Retire     ReleaseRetireAction
}

func (r ReleaseRunner) Release(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := LookupReleaseLease(ctx, r.Store, req, r.IsNotFound)
	if err != nil {
		return nil, err
	}
	return RetireReleaseLease(ctx, ReleaseRetireInput{
		Store:      r.Store,
		Locks:      r.Locks,
		Lease:      lease,
		IsNotFound: r.IsNotFound,
		Retire:     r.Retire,
	})
}
