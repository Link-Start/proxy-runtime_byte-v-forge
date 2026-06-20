package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

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
