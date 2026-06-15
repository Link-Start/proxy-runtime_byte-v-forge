package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) concurrencySlotRefreshRunner() leaseapp.RefreshConcurrencySlotRunner {
	return leaseapp.RefreshConcurrencySlotRunner{
		Store:      c.deps.store,
		Limiter:    c.deps.providerConcurrency,
		DefaultTTL: leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:  providerAccountConcurrencyTTLBuffer,
		Limit: func(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, policy *proxyruntimev1.ProxySessionPolicy) (uint32, error) {
			settings, err := c.deps.settings.load(ctx)
			if err != nil {
				return 0, err
			}
			return dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), nil
		},
	}
}
