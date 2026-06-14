package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
)

type leaseCoordinatorSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type leaseCoordinatorStore interface {
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

type leaseSessionProviderFactory interface {
	NewSessionProvider(accountproxy.Config) (provider.SessionProvider, error)
}

type leaseDataPlaneApplier interface {
	UpsertSessionRoute(context.Context, dataplane.SessionRoute) error
	DeleteSessionRoute(context.Context, dataplane.SessionRoute) error
}

type leaseLockManager interface {
	WithAccountLock(context.Context, string, leaseRuntimeLockFunc) error
	WithProviderAccountLock(context.Context, string, leaseRuntimeLockFunc) error
	WithSessionListenerAllocationLock(context.Context, leaseRuntimeLockFunc) error
}

type leaseRegistrySessionProviderFactory struct {
	registry *providerregistry.Registry
	client   *http.Client
}

func (f leaseRegistrySessionProviderFactory) NewSessionProvider(providerCfg accountproxy.Config) (provider.SessionProvider, error) {
	if f.registry == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return f.registry.NewSessionProvider(providerCfg, f.client)
}

type leaseListenerFunc func(context.Context, *runtimeSettingsFile, string, string) (config.EgressListener, error)
type leaseEndpointFunc func(config.EgressListener, string) (*proxyruntimev1.ProxyEndpoint, error)
type leaseAdvertisedHostFunc func(string, config.EgressListener) string
type leaseDialerProxyFunc func(context.Context, *runtimeSettingsFile, string) (string, map[string]string, error)
type leaseConnectionCleanupFunc func(context.Context, []string)

type leaseCoordinatorDependencies struct {
	cfg                     config.Config
	store                   leaseCoordinatorStore
	settings                leaseCoordinatorSettings
	clock                   leaseapp.Clock
	locks                   leaseLockManager
	dataPlane               leaseDataPlaneApplier
	dynamicIPSelector       *dynamicIPSelector
	sessionProviders        leaseSessionProviderFactory
	providerConcurrency     providerAccountConcurrencyLimiter
	logger                  leaseapp.Logger
	exitCheckCache          *proxyExitCheckCache
	leaseListener           leaseListenerFunc
	localListenerEndpoint   leaseEndpointFunc
	sessionAdvertisedHost   leaseAdvertisedHostFunc
	dynamicLeaseDialerProxy leaseDialerProxyFunc
	closeInUserConnections  leaseConnectionCleanupFunc
}

type leaseCoordinator struct {
	deps leaseCoordinatorDependencies
}

func newLeaseCoordinator(runtime *Runtime) leaseCoordinator {
	var store leaseCoordinatorStore
	if runtime.store != nil {
		store = runtime.store
	}
	return leaseCoordinator{deps: leaseCoordinatorDependencies{
		cfg:                     runtime.cfg,
		store:                   store,
		settings:                runtime.settings,
		clock:                   leaseapp.SystemClock{},
		locks:                   runtime.leaseLocks,
		dataPlane:               runtime.dataPlane,
		dynamicIPSelector:       runtime.dynamicIPSelector,
		sessionProviders:        leaseRegistrySessionProviderFactory{registry: runtime.accountProviders, client: runtime.providerHTTPClient},
		providerConcurrency:     runtime.providerConcurrency,
		logger:                  runtime.logger,
		exitCheckCache:          &runtime.exitCheckCache,
		leaseListener:           runtime.leaseListener,
		localListenerEndpoint:   runtime.localListenerEndpoint,
		sessionAdvertisedHost:   runtime.sessionAdvertisedHost,
		dynamicLeaseDialerProxy: runtime.dynamicLeaseDialerProxy,
		closeInUserConnections:  runtime.closeMihomoInUserConnections,
	}}
}

func (c leaseCoordinator) warn(message string, args ...any) {
	if c.deps.logger != nil {
		c.deps.logger.Warn(message, args...)
	}
}

func (c leaseCoordinator) now() time.Time {
	if c.deps.clock != nil {
		return c.deps.clock.Now()
	}
	return time.Now()
}

func (c leaseCoordinator) clearExitCheckCache() {
	if c.deps.exitCheckCache != nil {
		c.deps.exitCheckCache.clear()
	}
}

func (c leaseCoordinator) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if c.deps.closeInUserConnections != nil {
		c.deps.closeInUserConnections(ctx, usernames)
	}
}

func (c leaseCoordinator) activeLeaseByRequest(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if c.deps.store == nil {
		return nil, fmt.Errorf("lease store is required")
	}
	if sessionID != "" {
		return c.deps.store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return c.deps.store.ActiveLeaseFactByAccount(ctx, req.GetAccountId(), req.GetPurpose())
}

func (c leaseCoordinator) acquireProviderAccountConcurrencySlot(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (providerAccountConcurrencySlot, error) {
	return acquireProviderAccountConcurrencySlot(ctx, c.deps.providerConcurrency, account, limit, policy, holder, ttl)
}

func (c leaseCoordinator) releaseLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	accountID := strings.TrimSpace(lease.GetProviderAccountId())
	holder := strings.TrimSpace(leaseapp.ConcurrencyHolder(lease))
	if accountID == "" || holder == "" || c.deps.providerConcurrency == nil {
		return nil
	}
	return c.deps.providerConcurrency.Release(ctx, accountID, leaseapp.ConcurrencyPolicy(lease), holder)
}

func (c leaseCoordinator) refreshLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	accountID := strings.TrimSpace(lease.GetProviderAccountId())
	holder := strings.TrimSpace(leaseapp.ConcurrencyHolder(lease))
	if accountID == "" || holder == "" || c.deps.providerConcurrency == nil {
		return nil
	}
	account, err := c.deps.store.ProviderAccount(ctx, accountID)
	if err != nil {
		return err
	}
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return err
	}
	policy := leaseapp.ConcurrencyPolicy(lease)
	_, err = c.acquireProviderAccountConcurrencySlot(ctx, account, dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), policy, holder, leaseConcurrencySlotTTL(policy))
	return err
}

func (c leaseCoordinator) newSessionProvider(providerCfg accountproxy.Config) (provider.SessionProvider, error) {
	if c.deps.sessionProviders == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return c.deps.sessionProviders.NewSessionProvider(providerCfg)
}
