package lease

import (
	"context"
	"errors"
	"fmt"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func SaveActiveFact(ctx context.Context, store OrchestrationStore, input ActiveFactInput) (*proxygatewayv1.ProxyDynamicLease, error) {
	lease := NewActiveFact(input)
	return lease, store.SaveLeaseFact(ctx, lease)
}

func SaveReleaseCleanupFailure(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	MarkReleaseCleanupFailure(lease, routePending, providerPending, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveExpiredCleanupFailure(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease, routePending bool, providerPending bool, message string) error {
	MarkExpiredCleanupFailure(lease, routePending, providerPending, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveCleanupRetry(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease, message string) error {
	MarkCleanupRetry(lease, message)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveExpired(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease) error {
	MarkExpired(lease)
	return store.SaveLeaseFact(ctx, lease)
}

func SaveReleased(ctx context.Context, store OrchestrationStore, lease *proxygatewayv1.ProxyDynamicLease) error {
	MarkReleased(lease)
	return store.SaveLeaseFact(ctx, lease)
}

var ErrFinalLeaseConcurrencyRelease = errors.New("release provider account concurrency slot failed")

func SaveExpiredFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxygatewayv1.ProxyDynamicLease) error {
	if err := SaveExpired(ctx, store, lease); err != nil {
		return err
	}
	return releaseFinalLeaseConcurrencySlot(ctx, limiter, lease)
}

func SaveReleasedFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxygatewayv1.ProxyDynamicLease) error {
	if err := SaveReleased(ctx, store, lease); err != nil {
		return err
	}
	return releaseFinalLeaseConcurrencySlot(ctx, limiter, lease)
}

func releaseFinalLeaseConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, lease *proxygatewayv1.ProxyDynamicLease) error {
	if err := ReleaseLeaseConcurrencySlot(ctx, limiter, lease); err != nil {
		return fmt.Errorf("%w: %w", ErrFinalLeaseConcurrencyRelease, err)
	}
	return nil
}

func SaveCleanupProgress(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxygatewayv1.ProxyDynamicLease) error {
	if CleanupPending(lease) {
		if err := store.SaveLeaseFact(ctx, lease); err != nil {
			return err
		}
		return nil
	}
	switch CleanupFinalStatus(lease) {
	case CleanupFinalExpired:
		return SaveExpiredFinalLeaseState(ctx, store, limiter, lease)
	case CleanupFinalReleased:
		return SaveReleasedFinalLeaseState(ctx, store, limiter, lease)
	default:
		if err := store.SaveLeaseFact(ctx, lease); err != nil {
			return err
		}
		return nil
	}
}
