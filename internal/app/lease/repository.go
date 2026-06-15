package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type Repository interface {
	ListActiveLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListHistoryLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	LeaseFactByID(context.Context, string) (*proxyruntimev1.ProxyDynamicLease, error)
}

type Coordinator interface {
	Acquire(context.Context, string, *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error)
	Release(context.Context, *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error)
}

type Worker interface {
	RestoreActive(context.Context) error
	ExpireDue(context.Context) error
	CleanupPending(context.Context) error
	Cleanup(context.Context, *proxyruntimev1.ProxyDynamicLease) error
}
