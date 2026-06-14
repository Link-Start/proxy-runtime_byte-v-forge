package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type OrchestrationStore interface {
	ActiveLeaseFactBySession(context.Context, string, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
	ActiveLeaseFactByAccount(context.Context, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
	LatestLeaseFactByAccount(context.Context, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
	LeaseFactByID(context.Context, string) (*proxyruntimev1.ProxyDynamicLease, error)
	SaveLeaseFact(context.Context, *proxyruntimev1.ProxyDynamicLease) error
	CleanupPendingLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ExpiredActiveLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRestorableLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ProviderAccount(context.Context, string) (*proxyruntimev1.ProxyProviderAccount, error)
	ProviderConfig(context.Context, string) (accountproxy.Config, string, error)
}

type SessionProvider interface {
	Name() string
	CreateSession(context.Context, *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxySession, error)
	FetchSession(context.Context, *proxyruntimev1.ProxySession) ([]provider.Node, error)
	ReleaseSession(context.Context, *proxyruntimev1.ProxySession) error
}

type SessionProviderFactory interface {
	NewSessionProvider(accountproxy.Config) (SessionProvider, error)
}

type LocalService struct {
	Name     string
	Addr     string
	Protocol string
	Username string
	Password string
	Route    string
}

type SessionRoute struct {
	SessionID   string
	Listener    LocalService
	Pool        []provider.Node
	DialerProxy string
}

type DataPlaneApplier interface {
	UpsertSessionRoute(context.Context, SessionRoute) error
	DeleteSessionRoute(context.Context, SessionRoute) error
}

type LockFunc func(context.Context) error

type LockManager interface {
	WithAccountLock(context.Context, string, LockFunc) error
	WithProviderAccountLock(context.Context, string, LockFunc) error
	WithSessionListenerAllocationLock(context.Context, LockFunc) error
}

type IDGenerator interface {
	NewLeaseID() (string, error)
}
