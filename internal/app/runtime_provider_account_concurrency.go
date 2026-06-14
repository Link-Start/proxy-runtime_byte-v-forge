package app

import (
	"context"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	defaultDynamicProviderRotatingConcurrencyLimit uint32 = 10
	defaultDynamicProviderStickyConcurrencyLimit   uint32 = 2
	providerAccountConcurrencyTTLBuffer                   = 2 * time.Minute
)

func normalizeDynamicProviderRotatingConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return defaultDynamicProviderRotatingConcurrencyLimit
	}
	return value
}

func normalizeDynamicProviderStickyConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return defaultDynamicProviderStickyConcurrencyLimit
	}
	return value
}

func dynamicProviderInstanceConcurrencyLimit(provider dynamicIPProviderInstance, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	if providerAccountConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return normalizeDynamicProviderRotatingConcurrencyLimit(provider.rotatingConcurrencyLimit)
	}
	return normalizeDynamicProviderStickyConcurrencyLimit(provider.stickyConcurrencyLimit)
}

func dynamicProviderConcurrencyLimit(settings *runtimeSettingsFile, dynamicProviderID string, policy *proxyruntimev1.ProxySessionPolicy) uint32 {
	dynamicProviderID = runtimeSafeID(dynamicProviderID)
	for _, provider := range dynamicIPProviderInstances(settings) {
		if provider.dynamicProviderID == dynamicProviderID {
			return dynamicProviderInstanceConcurrencyLimit(provider, policy)
		}
	}
	if providerAccountConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return defaultDynamicProviderRotatingConcurrencyLimit
	}
	return defaultDynamicProviderStickyConcurrencyLimit
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

func (r *Runtime) acquireProviderAccountConcurrencySlot(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (providerAccountConcurrencySlot, error) {
	return acquireProviderAccountConcurrencySlot(ctx, r.providerConcurrency, account, limit, policy, holder, ttl)
}

func acquireProviderAccountConcurrencySlot(ctx context.Context, limiter providerAccountConcurrencyLimiter, account *proxyruntimev1.ProxyProviderAccount, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (providerAccountConcurrencySlot, error) {
	if limiter == nil {
		return noopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := limiter.Acquire(ctx, account.GetAccountId(), policy, limit, holder, ttl)
	if err != nil {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached: %w", account.GetAccountId(), providerAccountConcurrencyModeText(policy), err)
	}
	return slot, nil
}

func leaseConcurrencySlotTTL(policy *proxyruntimev1.ProxySessionPolicy) time.Duration {
	ttl := defaultDynamicIPStickyTTL
	if policy != nil && policy.GetStickyTtl() != nil && policy.GetStickyTtl().AsDuration() > 0 {
		ttl = policy.GetStickyTtl().AsDuration()
	}
	return ttl + providerAccountConcurrencyTTLBuffer
}
