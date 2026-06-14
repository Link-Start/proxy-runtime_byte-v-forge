package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func acquiredRouteApplyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrAcquiredRouteDataPlane):
		return unavailable("dataplane route apply failed", err)
	case errors.Is(err, leaseapp.ErrAcquiredRouteFactSave):
		return internalError("lease fact save failed", err)
	default:
		return err
	}
}
