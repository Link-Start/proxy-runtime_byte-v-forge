package app

import (
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func retryLeaseAcquireAttempt(err error) bool {
	return appcore.IsUnavailable(err) || appcore.IsFailedPrecondition(err)
}
