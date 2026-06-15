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
	return ReleaseLease(ctx, ReleaseInput{
		Store:      r.Store,
		Locks:      r.Locks,
		Request:    req,
		IsNotFound: r.IsNotFound,
		Retire:     r.Retire,
	})
}
