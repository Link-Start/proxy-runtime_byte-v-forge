package store

import (
	"context"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

type StoreCloser interface {
	Close()
}

type SecretStore interface {
	secretref.Writer
	secretref.Resolver
}

type RuntimeSettingsPersistence interface {
	LoadRuntimeSettings(context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
	SaveRuntimeSettings(context.Context, *proxyruntimev1.ProxyRuntimePersistentSettings) error
	LoadMihomoNativeSettings(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	SaveMihomoNativeSettings(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
}

type ProviderAccountStore interface {
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxyruntimev1.ProxyProviderAccount, error)
	ProviderAccountMutationState(context.Context, string) (ProviderAccountMutationState, error)
	ProviderConfig(context.Context, string) (accountproxy.Config, string, error)
	DefaultProviderAccountID(context.Context) (string, error)
}

type LeaseFactStore interface {
	SaveLeaseFact(context.Context, *proxyruntimev1.ProxyDynamicLease) error
	ListActiveLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListHistoryLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	RecentLeaseFacts(context.Context, time.Time, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ProviderAccountHasBlockingLease(context.Context, string) (bool, error)
	BlockingLeaseFactsByProviderAccount(context.Context, string, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
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
	RuntimeSettingsPersistence
	ProviderAccountStore
	LeaseFactStore
	SecretStore
	StoreCloser
}

func NewRuntimeStores(backend interface {
	RuntimeSettingsPersistence
	ProviderAccountStore
	LeaseFactStore
	SecretStore
	StoreCloser
}) *RuntimeStores {
	if backend == nil {
		return nil
	}
	return &RuntimeStores{
		RuntimeSettingsPersistence: backend,
		ProviderAccountStore:       backend,
		LeaseFactStore:             backend,
		SecretStore:                backend,
		StoreCloser:                backend,
	}
}

type ProviderAccountMutationState struct {
	ProviderID         string
	DynamicProviderID  string
	Username           string
	PasswordConfigured bool
	PasswordSecretRef  *commonv1.SecretRef
}
