package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) refreshLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if !leaseapp.NeedsConcurrencySlotRefresh(lease, c.deps.providerConcurrency) {
		return nil
	}
	settings, err := c.deps.settings.load(ctx)
	if err != nil {
		return err
	}
	policy := leaseapp.ConcurrencyPolicy(lease)
	return leaseapp.RefreshConcurrencySlot(ctx, leaseapp.RefreshConcurrencySlotInput{
		Store:      c.deps.store,
		Limiter:    c.deps.providerConcurrency,
		Lease:      lease,
		Limit:      dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy),
		DefaultTTL: leaseapp.DefaultDynamicIPStickyTTL,
		TTLBuffer:  providerAccountConcurrencyTTLBuffer,
	})
}
