package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	defaultProviderAccountRotatingConcurrencyLimit uint32 = 10
	defaultProviderAccountStickyConcurrencyLimit   uint32 = 2
	providerAccountConcurrencyTTLBuffer                   = 2 * time.Minute
)

func normalizeProviderAccountRotatingConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return defaultProviderAccountRotatingConcurrencyLimit
	}
	return value
}

func normalizeProviderAccountStickyConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return defaultProviderAccountStickyConcurrencyLimit
	}
	return value
}

func storedProviderAccountConcurrencyLimit(value int64) uint32 {
	if value <= 0 {
		return 0
	}
	if value > int64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(value)
}

func providerAccountConcurrencyLimit(account *proxyruntimev1.ProxyProviderAccount, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	if providerAccountConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return normalizeProviderAccountRotatingConcurrencyLimit(account.GetRotatingConcurrencyLimit())
	}
	return normalizeProviderAccountStickyConcurrencyLimit(account.GetStickyConcurrencyLimit())
}

func providerAccountConcurrencyMode(policy *proxyruntimev1.ProxySessionPolicy) proxyruntimev1.ProxySessionMode {
	if policy == nil {
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
	if policy.GetMode() == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING || policy.GetRotationMode() == proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST {
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	}
	return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
}

func providerAccountConcurrencyModeText(policy *proxyruntimev1.ProxySessionPolicy) string {
	if providerAccountConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return "rotating"
	}
	return "sticky"
}

func (r *Runtime) providerAccountConcurrencyAvailable(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, policy *proxyruntimev1.ProxySessionPolicy, holder string) (bool, error) {
	if r.providerConcurrency == nil {
		return true, nil
	}
	return r.providerConcurrency.Available(ctx, account.GetAccountId(), policy, providerAccountConcurrencyLimit(account, policy), holder)
}

func (r *Runtime) acquireProviderAccountConcurrencySlot(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (providerAccountConcurrencySlot, error) {
	if r.providerConcurrency == nil {
		return noopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := r.providerConcurrency.Acquire(ctx, account.GetAccountId(), policy, providerAccountConcurrencyLimit(account, policy), holder, ttl)
	if err != nil {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached: %w", account.GetAccountId(), providerAccountConcurrencyModeText(policy), err)
	}
	return slot, nil
}

func (r *Runtime) releaseLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	accountID := strings.TrimSpace(lease.GetProviderAccountId())
	holder := leaseConcurrencyHolder(lease)
	if accountID == "" || holder == "" {
		return nil
	}
	if r.providerConcurrency == nil {
		return nil
	}
	return r.providerConcurrency.Release(ctx, accountID, leaseConcurrencyPolicy(lease), holder)
}

func (r *Runtime) refreshLeaseConcurrencySlot(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	accountID := strings.TrimSpace(lease.GetProviderAccountId())
	holder := leaseConcurrencyHolder(lease)
	if accountID == "" || holder == "" {
		return nil
	}
	account, err := r.store.ProviderAccount(ctx, accountID)
	if err != nil {
		return err
	}
	_, err = r.acquireProviderAccountConcurrencySlot(ctx, account, leaseConcurrencyPolicy(lease), holder, leaseConcurrencySlotTTL(leaseConcurrencyPolicy(lease)))
	return err
}

func leaseConcurrencyPolicy(lease *proxyruntimev1.ProxyDynamicLease) *proxyruntimev1.ProxySessionPolicy {
	if lease == nil {
		return nil
	}
	if policy := lease.GetSession().GetPolicy(); policy != nil {
		return policy
	}
	return &proxyruntimev1.ProxySessionPolicy{RotationMode: lease.GetEgress().GetRotationMode()}
}

func leaseConcurrencyHolder(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if holder := strings.TrimSpace(lease.GetEgress().GetLabels()["provider_account_concurrency_holder"]); holder != "" {
		return holder
	}
	if holder := strings.TrimSpace(lease.GetSession().GetLabels()["provider_account_concurrency_holder"]); holder != "" {
		return holder
	}
	if leaseID := strings.TrimSpace(lease.GetLeaseId()); leaseID != "" {
		return "lease:" + leaseID
	}
	return ""
}

func leaseConcurrencySlotTTL(policy *proxyruntimev1.ProxySessionPolicy) time.Duration {
	ttl := defaultDynamicIPStickyTTL
	if policy != nil && policy.GetStickyTtl() != nil && policy.GetStickyTtl().AsDuration() > 0 {
		ttl = policy.GetStickyTtl().AsDuration()
	}
	return ttl + providerAccountConcurrencyTTLBuffer
}

func dynamicProfileConcurrencySlotTTL(refreshInterval time.Duration, policy *proxyruntimev1.ProxySessionPolicy) time.Duration {
	ttl := leaseConcurrencySlotTTL(policy)
	minTTL := defaultProviderAccountConcurrencySlotTTL
	if refreshInterval > 0 {
		minTTL = refreshInterval*2 + time.Minute
	}
	if ttl < minTTL {
		return minTTL
	}
	return ttl
}

func dynamicProfileConcurrencyHolder(sessionKey string) string {
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return ""
	}
	return "dynamic-profile:" + sessionKey
}
