package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func acquireAttemptSlotError(stage leaseapp.AcquireAttemptSlotStage, err error) error {
	if stage == leaseapp.AcquireAttemptSlotConcurrencyAcquire {
		return failedPrecondition("provider account concurrency limit reached", err)
	}
	return err
}
