package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func acquireAttemptSlotError(err error) error {
	if errors.Is(err, leaseapp.ErrAcquireAttemptConcurrencyLimit) {
		return failedPrecondition("provider account concurrency limit reached", err)
	}
	return err
}
