package lease

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type Repository interface {
	ListActiveLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ListHistoryLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	LeaseFactByID(context.Context, string) (*proxygatewayv1.ProxyDynamicLease, error)
}

type Coordinator interface {
	Acquire(context.Context, string, *proxygatewayv1.AcquireProxyLeaseRequest) (*proxygatewayv1.ProxyDynamicLease, error)
	Release(context.Context, *proxygatewayv1.ReleaseProxyLeaseRequest) (*proxygatewayv1.ProxyDynamicLease, error)
}

type Worker interface {
	RestoreActive(context.Context) error
	ExpireDue(context.Context) error
	CleanupPending(context.Context) error
	Cleanup(context.Context, *proxygatewayv1.ProxyDynamicLease) error
}
