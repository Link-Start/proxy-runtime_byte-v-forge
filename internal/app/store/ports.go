package store

import (
	"context"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/common/v1"
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"
)

type StoreCloser interface {
	Close()
}

type SecretStore interface {
	secretref.Writer
	secretref.Resolver
}

type RuntimeSettingsPersistence interface {
	LoadRuntimeSettings(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	SaveRuntimeSettings(context.Context, *proxygatewayv1.ProxyGatewayPersistentSettings) error
	LoadMihomoNativeSettings(context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error)
	SaveMihomoNativeSettings(context.Context, *proxygatewayv1.ProxyGatewayMihomoNativeConfig) error
}

type ProviderAccountStore interface {
	ListProviderAccounts(context.Context) ([]*proxygatewayv1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxygatewayv1.UpsertProxyProviderAccountRequest) (*proxygatewayv1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxygatewayv1.ProxyProviderAccount, error)
	ProviderAccountMutationState(context.Context, string) (ProviderAccountMutationState, error)
	ProviderConfig(context.Context, string) (accountproxy.Config, string, error)
	DefaultProviderAccountID(context.Context) (string, error)
}

type LeaseFactStore interface {
	SaveLeaseFact(context.Context, *proxygatewayv1.ProxyDynamicLease) error
	ListActiveLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ListHistoryLeaseFacts(context.Context, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	RecentLeaseFacts(context.Context, time.Time, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ProviderAccountHasBlockingLease(context.Context, string) (bool, error)
	BlockingLeaseFactsByProviderAccount(context.Context, string, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	CleanupPendingLeaseFacts(context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ListRestorableLeaseFacts(context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ExpiredActiveLeaseFacts(context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error)
	LeaseFactByID(context.Context, string) (*proxygatewayv1.ProxyDynamicLease, error)
	ActiveLeaseFact(context.Context, string) (*proxygatewayv1.ProxyDynamicLease, error)
	ActiveLeaseFactBySession(context.Context, string, string, string) (*proxygatewayv1.ProxyDynamicLease, error)
	ActiveLeaseFactByAccount(context.Context, string, string) (*proxygatewayv1.ProxyDynamicLease, error)
	LatestLeaseFactByAccount(context.Context, string, string) (*proxygatewayv1.ProxyDynamicLease, error)
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
