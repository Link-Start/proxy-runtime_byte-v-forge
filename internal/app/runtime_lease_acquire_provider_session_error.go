package app

import (
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

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
