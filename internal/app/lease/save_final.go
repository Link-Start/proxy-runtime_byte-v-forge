package lease

import (
	"context"
	"errors"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var (
	ErrFinalLeaseStateUnsupported   = errors.New("final lease state is unsupported")
	ErrFinalLeaseConcurrencyRelease = errors.New("release provider account concurrency slot failed")
)

type FinalLeaseState int

const (
	FinalLeaseStateExpired FinalLeaseState = iota + 1
	FinalLeaseStateReleased
)

func SaveFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease, state FinalLeaseState) error {
	if err := saveFinalLeaseState(ctx, store, lease, state); err != nil {
		return err
	}
	if err := ReleaseLeaseConcurrencySlot(ctx, limiter, lease); err != nil {
		return fmt.Errorf("%w: %w", ErrFinalLeaseConcurrencyRelease, err)
	}
	return nil
}

func saveFinalLeaseState(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease, state FinalLeaseState) error {
	switch state {
	case FinalLeaseStateExpired:
		return SaveExpired(ctx, store, lease)
	case FinalLeaseStateReleased:
		return SaveReleased(ctx, store, lease)
	default:
		return ErrFinalLeaseStateUnsupported
	}
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
		return SaveFinalLeaseState(ctx, store, limiter, lease, FinalLeaseStateExpired)
	case CleanupFinalReleased:
		return SaveFinalLeaseState(ctx, store, limiter, lease, FinalLeaseStateReleased)
	default:
		if err := store.SaveLeaseFact(ctx, lease); err != nil {
			return err
		}
		return nil
	}
}
