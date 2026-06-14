package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireRestoreLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease, settings *runtimeSettingsFile, providerAccount *proxyruntimev1.ProxyProviderAccount) (leaseapp.ProviderAccountConcurrencySlot, error) {
	holder := leaseapp.ConcurrencyHolder(lease)
	policy := leaseapp.ConcurrencyPolicy(lease)
	return c.acquireProviderAccountConcurrencySlot(ctx, providerAccount, dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), policy, holder, leaseapp.ConcurrencySlotTTL(policy, defaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
}

func releaseRestoreLeaseConcurrencySlotUnlessKept(ctx context.Context, slot leaseapp.ProviderAccountConcurrencySlot, keep *bool) {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), leaseRestoreSlotReleaseTimeout)
	defer cancel()
	_ = leaseapp.ReleaseConcurrencySlotUnlessKept(releaseCtx, slot, keep != nil && *keep)
}
