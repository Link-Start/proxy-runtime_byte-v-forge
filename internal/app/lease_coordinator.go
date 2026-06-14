package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

type leaseCoordinatorSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type leaseListenerFunc func(context.Context, *runtimeSettingsFile, string, string) (config.EgressListener, error)
type leaseEndpointFunc func(config.EgressListener, string) (*proxyruntimev1.ProxyEndpoint, error)
type leaseAdvertisedHostFunc func(string, config.EgressListener) string
type leaseDialerProxyFunc func(context.Context, *runtimeSettingsFile, string) (string, map[string]string, error)
type leaseConnectionCleanupFunc func(context.Context, []string)

type leaseCoordinatorDependencies struct {
	cfg                     config.Config
	store                   leaseapp.OrchestrationStore
	settings                leaseCoordinatorSettings
	clock                   leaseapp.Clock
	ids                     leaseapp.IDGenerator
	locks                   leaseapp.LockManager
	dataPlane               leaseapp.DataPlaneApplier
	dynamicIPSelector       *dynamicIPSelector
	sessionProviders        leaseapp.SessionProviderFactory
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

func newLeaseCoordinator(deps leaseCoordinatorDependencies) leaseCoordinator {
	if deps.clock == nil {
		deps.clock = leaseapp.SystemClock{}
	}
	if deps.ids == nil {
		deps.ids = randomLeaseIDGenerator{byteLength: leaseIDByteLength}
	}
	return leaseCoordinator{deps: deps}
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

func (c leaseCoordinator) newLeaseID() (string, error) {
	if c.deps.ids == nil {
		return "", fmt.Errorf("lease id generator is required")
	}
	return c.deps.ids.NewLeaseID()
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
	_, err = c.acquireProviderAccountConcurrencySlot(ctx, account, dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), policy, holder, leaseapp.ConcurrencySlotTTL(policy, defaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	return err
}

func (c leaseCoordinator) newSessionProvider(providerCfg accountproxy.Config) (provider.SessionProvider, error) {
	if c.deps.sessionProviders == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return c.deps.sessionProviders.NewSessionProvider(providerCfg)
}
