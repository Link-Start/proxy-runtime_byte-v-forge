package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func leaseProfilePolicyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProfileDynamicIPNotConfigured), errors.Is(err, leaseapp.ErrProfileLeaseRequiresSticky):
		return appcore.FailedPrecondition(err.Error(), err)
	case errors.Is(err, leaseapp.ErrRequestRequiresStickyDynamicIP):
		return appcore.InvalidArgument(err.Error(), err)
	default:
		return err
	}
}
