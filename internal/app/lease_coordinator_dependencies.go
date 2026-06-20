package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

type leaseCoordinatorSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type leaseListenerFunc func(context.Context, *runtimeSettingsFile, string, string) (leaseapp.Listener, error)
type leaseEndpointFunc func(leaseapp.Listener, string) (*proxyruntimev1.ProxyEndpoint, error)
type leaseAdvertisedHostFunc func(string, leaseapp.Listener) string
type leaseDialerProxyFunc func(context.Context, *runtimeSettingsFile, string) (string, map[string]string, error)
type leaseConnectionCleanupFunc func(context.Context, []string)

type leaseCoordinatorDependencies struct {
	cfg                     config.Config
	store                   leaseapp.OrchestrationStore
	settings                leaseCoordinatorSettings
	clock                   clock.Clock
	ids                     leaseapp.IDGenerator
	locks                   leaseapp.LockManager
	dataPlane               leaseapp.DataPlaneApplier
	dynamicIPSelector       *dynamic.IPSelector
	sessionProviders        leaseapp.SessionProviderFactory
	providerConcurrency     leaseapp.ProviderAccountConcurrencyLimiter
	logger                  leaseapp.Logger
	exitCheckCache          *proxycheck.ExitCheckCache
	leaseListener           leaseListenerFunc
	localListenerEndpoint   leaseEndpointFunc
	sessionAdvertisedHost   leaseAdvertisedHostFunc
	dynamicLeaseDialerProxy leaseDialerProxyFunc
	closeInUserConnections  leaseConnectionCleanupFunc
}
