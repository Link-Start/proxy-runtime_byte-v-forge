package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func acquiredRouteApplyError(stage leaseapp.AcquiredRouteApplyStage, err error) error {
	switch stage {
	case leaseapp.AcquiredRouteApplyDataPlane:
		return unavailable("dataplane route apply failed", err)
	case leaseapp.AcquiredRouteApplyFactSave:
		return internalError("lease fact save failed", err)
	default:
		return err
	}
}
