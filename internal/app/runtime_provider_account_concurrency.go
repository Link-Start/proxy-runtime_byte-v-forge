package app

import (
	"context"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
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
	if leaseapp.ConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
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
	if leaseapp.ConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return defaultDynamicProviderRotatingConcurrencyLimit
	}
	return defaultDynamicProviderStickyConcurrencyLimit
}

func acquireProviderAccountConcurrencySlot(ctx context.Context, limiter leaseapp.ProviderAccountConcurrencyLimiter, account *proxyruntimev1.ProxyProviderAccount, limit uint32, policy *proxyruntimev1.ProxySessionPolicy, holder string, ttl time.Duration) (leaseapp.ProviderAccountConcurrencySlot, error) {
	if limiter == nil {
		return leaseapp.NoopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := limiter.Acquire(ctx, account.GetAccountId(), policy, limit, holder, ttl)
	if err != nil {
		return nil, fmt.Errorf("provider account %q %s concurrency limit reached: %w", account.GetAccountId(), leaseapp.ConcurrencyModeText(policy), err)
	}
	return slot, nil
}
