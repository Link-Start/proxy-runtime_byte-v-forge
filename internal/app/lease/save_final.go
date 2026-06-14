package lease

import (
	"context"
	"errors"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrFinalLeaseConcurrencyRelease = errors.New("release provider account concurrency slot failed")

func SaveExpiredFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
	if err := SaveExpired(ctx, store, lease); err != nil {
		return err
	}
	return releaseFinalLeaseConcurrencySlot(ctx, limiter, lease)
}

func SaveReleasedFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
	if err := SaveReleased(ctx, store, lease); err != nil {
		return err
	}
	return releaseFinalLeaseConcurrencySlot(ctx, limiter, lease)
}

func releaseFinalLeaseConcurrencySlot(ctx context.Context, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
	if err := ReleaseLeaseConcurrencySlot(ctx, limiter, lease); err != nil {
		return fmt.Errorf("%w: %w", ErrFinalLeaseConcurrencyRelease, err)
	}
	return nil
}

func SaveCleanupProgress(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease) error {
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
