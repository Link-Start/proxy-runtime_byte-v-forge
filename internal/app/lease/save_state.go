package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func SaveActiveFact(ctx context.Context, store OrchestrationStore, input ActiveFactInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease := NewActiveFact(input)
	return lease, store.SaveLeaseFact(ctx, lease)
}

func SaveReleaseCleanupFailure(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	MarkReleaseCleanupFailure(lease, routePending, providerPending, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveExpiredCleanupFailure(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	MarkExpiredCleanupFailure(lease, routePending, providerPending, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveCleanupRetry(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease, message string) error {
	MarkCleanupRetry(lease, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveExpired(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease) error {
	MarkExpired(lease)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveReleased(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease) error {
	MarkReleased(lease)
	return store.SaveLeaseFact(ctx, lease)
}
