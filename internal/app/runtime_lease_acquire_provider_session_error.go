package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func providerSessionAcquireError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProviderSessionFactory):
		return invalidArgument("provider account configuration is invalid", err)
	case errors.Is(err, leaseapp.ErrProviderSessionCreate):
		return unavailable("provider session create failed", err)
	case errors.Is(err, leaseapp.ErrProviderSessionFetch):
		return unavailable("provider session fetch failed", err)
	default:
		return err
	}
}
