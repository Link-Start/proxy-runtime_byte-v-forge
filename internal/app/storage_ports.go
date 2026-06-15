package app

import (
	"context"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

type storeCloser interface {
	Close()
}

type secretStore interface {
	secretref.Writer
	secretref.Resolver
}

type runtimeSettingsPersistence interface {
	LoadRuntimeSettings(context.Context) (*runtimeSettingsFile, error)
	SaveRuntimeSettings(context.Context, *runtimeSettingsFile) error
	LoadMihomoNativeSettings(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	SaveMihomoNativeSettings(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
}

type providerAccountStore interface {
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxyruntimev1.ProxyProviderAccount, error)
	ProviderAccountMutationState(context.Context, string) (providerAccountMutationState, error)
	ProviderConfig(context.Context, string) (accountproxy.Config, string, error)
	DefaultProviderAccountID(context.Context) (string, error)
}

type leaseFactStore interface {
	SaveLeaseFact(context.Context, *proxyruntimev1.ProxyDynamicLease) error
	ListActiveLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListHistoryLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
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

type RuntimeStores struct {
	runtimeSettingsPersistence
	providerAccountStore
	leaseFactStore
	secretStore
	storeCloser
}

func newRuntimeStores(store interface {
	runtimeSettingsPersistence
	providerAccountStore
	leaseFactStore
	secretStore
	storeCloser
}) *RuntimeStores {
	if store == nil {
		return nil
	}
	return &RuntimeStores{
		runtimeSettingsPersistence: store,
		providerAccountStore:       store,
		leaseFactStore:             store,
		secretStore:                store,
		storeCloser:                store,
	}
}

type providerAccountMutationState struct {
	ProviderID         string
	DynamicProviderID  string
	Username           string
	PasswordConfigured bool
	PasswordSecretRef  *commonv1.SecretRef
}
