package app

import (
	"context"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

type controlStore interface {
	secretref.Writer
	secretref.Resolver
	Close()
	LoadRuntimeSettings(context.Context) (*runtimeSettingsFile, error)
	SaveRuntimeSettings(context.Context, *runtimeSettingsFile) error
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxyruntimev1.ProxyProviderAccount, error)
	ProviderAccountMutationState(context.Context, string) (providerAccountMutationState, error)
	ProviderConfig(context.Context, string) (accountproxy.Config, string, error)
	DefaultProviderAccountID(context.Context) (string, error)
	SaveLeaseFact(context.Context, *proxyruntimev1.ProxyDynamicLease) error
	ListLeaseFacts(context.Context, bool) ([]*proxyruntimev1.ProxyDynamicLease, error)
	RecentLeaseFacts(context.Context, time.Time, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ProviderAccountHasBlockingLease(context.Context, string) (bool, error)
	BlockingLeaseFactsByProviderAccount(context.Context, string) ([]*proxyruntimev1.ProxyDynamicLease, error)
	CleanupPendingLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRestorableLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ExpiredActiveLeaseFacts(context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error)
	LeaseFactByID(context.Context, string) (*proxyruntimev1.ProxyDynamicLease, error)
	ActiveLeaseFact(context.Context, string) (*proxyruntimev1.ProxyDynamicLease, error)
	ActiveLeaseFactBySession(context.Context, string, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
	ActiveLeaseFactByAccount(context.Context, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
	LatestLeaseFactByAccount(context.Context, string, string) (*proxyruntimev1.ProxyDynamicLease, error)
}

type providerAccountMutationState struct {
	ProviderID         string
	DynamicProviderID  string
	Username           string
	PasswordConfigured bool
	PasswordSecretRef  *commonv1.SecretRef
}
