package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func acquireAttemptSlotError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrAcquireAttemptLeaseID):
		return appcore.InternalError("generate lease id", err)
	case errors.Is(err, leaseapp.ErrAcquireAttemptConcurrencyLimit):
		return appcore.FailedPrecondition("provider account concurrency limit reached", err)
	default:
		return err
	}
}
