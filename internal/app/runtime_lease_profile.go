package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func leaseProfilePolicyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProfileDynamicIPNotConfigured), errors.Is(err, leaseapp.ErrProfileLeaseRequiresSticky):
		return failedPrecondition(err.Error(), err)
	case errors.Is(err, leaseapp.ErrRequestRequiresStickyDynamicIP):
		return invalidArgument(err.Error(), err)
	default:
		return err
	}
}
