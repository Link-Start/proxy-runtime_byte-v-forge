package app

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
)

type leaseConcurrencySlotRefreshRunnerFactory struct {
	deps leaseCoordinatorDependencies
}

func (f leaseConcurrencySlotRefreshRunnerFactory) New() leaseapp.RefreshConcurrencySlotRunner {
	return leaseapp.RefreshConcurrencySlotRunner{
		Store:      f.deps.store,
		Limiter:    f.deps.providerConcurrency,
		DefaultTTL: kernel.DefaultDynamicIPStickyTTL,
		TTLBuffer:  providerAccountConcurrencyTTLBuffer,
		Limit:      f.limit,
	}
}

func (f leaseConcurrencySlotRefreshRunnerFactory) limit(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease, policy *proxygatewayv1.ProxySessionPolicy) (uint32, error) {
	settings, err := f.deps.settings.Load(ctx)
	if err != nil {
		return 0, err
	}
	return dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), nil
}

func (c leaseCoordinator) concurrencySlotRefreshRunner() leaseapp.RefreshConcurrencySlotRunner {
	return leaseConcurrencySlotRefreshRunnerFactory{deps: c.deps}.New()
}
