package app

import (
	"errors"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
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

func providerSessionAcquireError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProviderSessionFactory):
		return appcore.InvalidArgument("provider account configuration is invalid", err)
	case errors.Is(err, leaseapp.ErrProviderSessionCreate):
		return appcore.Unavailable("provider session create failed", err)
	case errors.Is(err, leaseapp.ErrProviderSessionFetch):
		return appcore.Unavailable("provider session fetch failed", err)
	default:
		return err
	}
}

func retryLeaseAcquireAttempt(err error) bool {
	return appcore.IsUnavailable(err) || appcore.IsFailedPrecondition(err)
}

func acquiredRouteApplyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrAcquiredRouteDataPlane):
		return appcore.Unavailable("dataplane route apply failed", err)
	case errors.Is(err, leaseapp.ErrAcquiredRouteFactSave):
		return appcore.InternalError("lease fact save failed", err)
	default:
		return err
	}
}
