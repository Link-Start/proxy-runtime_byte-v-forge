package app

import (
	"errors"
)

func retryLeaseAcquireAttempt(err error) bool {
	var appErr *appError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.code == errCodeUnavailable || appErr.code == errCodeFailedPrecondition
}
