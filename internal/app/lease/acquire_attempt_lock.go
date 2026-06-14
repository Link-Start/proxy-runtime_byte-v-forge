package lease

import (
	"context"
	"errors"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAcquireAttemptActionRequired = errors.New("acquire attempt action is required")

type AcquireAttemptAction func(context.Context) (*proxyruntimev1.ProxyDynamicLease, error)

type LockedAcquireAttemptInput struct {
	Locks             LockManager
	ProviderAccountID string
	ConcurrencySlot   ProviderAccountConcurrencySlot
	ReleaseTimeout    time.Duration
	Action            AcquireAttemptAction
}

func RunLockedAcquireAttempt(ctx context.Context, input LockedAcquireAttemptInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	if input.Action == nil {
		return nil, ErrAcquireAttemptActionRequired
	}
	keepConcurrencySlot := false
	defer func() {
		releaseCtx := context.WithoutCancel(ctx)
		if input.ReleaseTimeout > 0 {
			var cancel context.CancelFunc
			releaseCtx, cancel = context.WithTimeout(releaseCtx, input.ReleaseTimeout)
			defer cancel()
		}
		_ = ReleaseConcurrencySlotUnlessKept(releaseCtx, input.ConcurrencySlot, keepConcurrencySlot)
	}()
	var lease *proxyruntimev1.ProxyDynamicLease
	err := WithProviderAccountLock(ctx, input.Locks, input.ProviderAccountID, func(ctx context.Context) error {
		var err error
		lease, err = input.Action(ctx)
		if err == nil {
			keepConcurrencySlot = true
		}
		return err
	})
	return lease, err
}
