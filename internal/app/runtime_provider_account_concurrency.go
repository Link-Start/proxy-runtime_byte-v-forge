package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
	if r.providerConcurrency == nil {
		return noopProviderAccountConcurrencySlot{}, nil
	}
	slot, err := r.providerConcurrency.Acquire(ctx, account.GetAccountId(), policy, limit, holder, ttl)
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
	settings, err := r.settings.load(ctx)
	if err != nil {
		return err
	}
	policy := leaseConcurrencyPolicy(lease)
	_, err = r.acquireProviderAccountConcurrencySlot(ctx, account, dynamicProviderConcurrencyLimit(settings, leaseDynamicProviderID(lease), policy), policy, holder, leaseConcurrencySlotTTL(policy))
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

func leaseDynamicProviderID(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetEgress().GetLabels()["dynamic_provider_id"]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetSession().GetLabels()["dynamic_provider_id"]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if endpoint := lease.GetSelectionPlan().GetSelectedEndpoint(); endpoint != nil {
		return endpoint.GetDynamicProviderId()
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
