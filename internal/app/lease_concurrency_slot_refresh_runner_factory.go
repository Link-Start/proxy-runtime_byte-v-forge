package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type leaseConcurrencySlotRefreshRunnerFactory struct {
	deps leaseCoordinatorDependencies
}

func (f leaseConcurrencySlotRefreshRunnerFactory) New() leaseapp.RefreshConcurrencySlotRunner {
	return leaseapp.RefreshConcurrencySlotRunner{
		Store:      f.deps.store,
		Limiter:    f.deps.providerConcurrency,
		DefaultTTL: leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:  providerAccountConcurrencyTTLBuffer,
		Limit:      f.limit,
	}
}

func (f leaseConcurrencySlotRefreshRunnerFactory) limit(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, policy *proxyruntimev1.ProxySessionPolicy) (uint32, error) {
	settings, err := f.deps.settings.load(ctx)
	if err != nil {
		return 0, err
	}
	return dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), nil
}

func (c leaseCoordinator) concurrencySlotRefreshRunner() leaseapp.RefreshConcurrencySlotRunner {
	return leaseConcurrencySlotRefreshRunnerFactory{deps: c.deps}.New()
}
