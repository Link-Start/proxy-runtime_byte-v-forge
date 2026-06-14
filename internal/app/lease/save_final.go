package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrFinalLeaseStateUnsupported = errors.New("final lease state is unsupported")

type FinalLeaseState int

const (
	FinalLeaseStateExpired FinalLeaseState = iota + 1
	FinalLeaseStateReleased
)

type FinalLeaseSaveStage int

const (
	FinalLeaseSaveNoError FinalLeaseSaveStage = iota
	FinalLeaseSaveState
	FinalLeaseSaveConcurrencyRelease
)

func SaveFinalLeaseState(ctx context.Context, store OrchestrationStore, limiter ProviderAccountConcurrencyLimiter, lease *proxyruntimev1.ProxyDynamicLease, state FinalLeaseState) (FinalLeaseSaveStage, error) {
	if err := saveFinalLeaseState(ctx, store, lease, state); err != nil {
		return FinalLeaseSaveState, err
	}
	if err := ReleaseLeaseConcurrencySlot(ctx, limiter, lease); err != nil {
		return FinalLeaseSaveConcurrencyRelease, err
	}
	return FinalLeaseSaveNoError, nil
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
