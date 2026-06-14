package lease

import (
	"context"
	"errors"
	"time"
)

var ErrTemporaryConcurrencyActionRequired = errors.New("temporary concurrency action is required")

type TemporaryConcurrencyAction func(context.Context) error

type TemporaryConcurrencySlotInput struct {
	Slot           ProviderAccountConcurrencySlot
	ReleaseTimeout time.Duration
	Action         TemporaryConcurrencyAction
}

func RunTemporaryConcurrencySlot(ctx context.Context, input TemporaryConcurrencySlotInput) error {
	if input.Action == nil {
		return ErrTemporaryConcurrencyActionRequired
	}
	keep := false
	defer func() {
		releaseCtx := context.WithoutCancel(ctx)
		if input.ReleaseTimeout > 0 {
			var cancel context.CancelFunc
			releaseCtx, cancel = context.WithTimeout(releaseCtx, input.ReleaseTimeout)
			defer cancel()
		}
		_ = ReleaseConcurrencySlotUnlessKept(releaseCtx, input.Slot, keep)
	}()
	if err := input.Action(ctx); err != nil {
		return err
	}
	keep = true
	return nil
}
