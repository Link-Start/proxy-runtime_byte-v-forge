package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) releaseLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	return leaseapp.ReleaseProviderAccountConcurrencySlot(ctx, c.deps.providerConcurrency, lease.GetProviderAccountId(), leaseapp.ConcurrencyPolicy(lease), leaseapp.ConcurrencyHolder(lease))
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
	_, err = leaseapp.AcquireProviderAccountConcurrencySlot(ctx, c.deps.providerConcurrency, account.GetAccountId(), dynamicProviderConcurrencyLimit(settings, leaseapp.DynamicProviderID(lease), policy), policy, holder, leaseapp.ConcurrencySlotTTL(policy, leaseapp.DefaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	return err
}
