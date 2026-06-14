package app

import (
	"context"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) acquireProviderAccountConcurrencySlot(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (leaseapp.ProviderAccountConcurrencySlot, error) {
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
