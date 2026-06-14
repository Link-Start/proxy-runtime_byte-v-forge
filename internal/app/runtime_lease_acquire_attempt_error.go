package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func acquireAttemptSlotError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrAcquireAttemptLeaseID):
		return internalError("generate lease id", err)
	case errors.Is(err, leaseapp.ErrAcquireAttemptConcurrencyLimit):
		return failedPrecondition("provider account concurrency limit reached", err)
	default:
		return err
	}
}
