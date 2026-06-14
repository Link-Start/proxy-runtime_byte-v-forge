package app

import (
	"errors"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
)

type appErrorCode string

const (
	errCodeInvalidArgument    appErrorCode = "invalid_argument"
	errCodeNotFound           appErrorCode = "not_found"
	errCodeFailedPrecondition appErrorCode = "failed_precondition"
	errCodeResourceExhausted  appErrorCode = "resource_exhausted"
	errCodeUnavailable        appErrorCode = "unavailable"
	errCodeInternal           appErrorCode = "internal"
)

type appError struct {
	code    appErrorCode
	message string
	cause   error
}

func (e *appError) Error() string { return e.message }

func (e *appError) Unwrap() error { return e.cause }

func newAppError(code appErrorCode, message string, cause error) error {
	message = strings.TrimSpace(message)
	if message == "" && cause != nil {
		message = cause.Error()
	}
	if message == "" {
		message = "request failed"
	}
	return &appError{code: code, message: message, cause: cause}
}

func invalidArgument(message string, cause error) error {
	return newAppError(errCodeInvalidArgument, message, cause)
}

func failedPrecondition(message string, cause error) error {
	return newAppError(errCodeFailedPrecondition, message, cause)
}

func resourceExhausted(message string, cause error) error {
	return newAppError(errCodeResourceExhausted, message, cause)
}

func unavailable(message string, cause error) error {
	return newAppError(errCodeUnavailable, message, cause)
}

func internalError(message string, cause error) error {
	return newAppError(errCodeInternal, message, cause)
}

func httpErrorDetails(err error, fallbackStatus int) (int, codes.Code, string) {
	var appErr *appError
	if errors.As(err, &appErr) {
		switch appErr.code {
		case errCodeInvalidArgument:
			return http.StatusBadRequest, codes.InvalidArgument, appErr.message
		case errCodeNotFound:
			return http.StatusNotFound, codes.NotFound, appErr.message
		case errCodeFailedPrecondition:
			return http.StatusPreconditionFailed, codes.FailedPrecondition, appErr.message
		case errCodeResourceExhausted:
			return http.StatusRequestEntityTooLarge, codes.ResourceExhausted, appErr.message
		case errCodeUnavailable:
			return http.StatusBadGateway, codes.Unavailable, appErr.message
		case errCodeInternal:
			return http.StatusInternalServerError, codes.Internal, "internal server error"
		default:
			return http.StatusInternalServerError, codes.Internal, "internal server error"
		}
	}
	switch fallbackStatus {
	case http.StatusUnauthorized:
		return fallbackStatus, codes.Unauthenticated, publicErrorMessage(err, fallbackStatus)
	case http.StatusBadRequest:
		return fallbackStatus, codes.InvalidArgument, publicErrorMessage(err, fallbackStatus)
	case http.StatusNotFound:
		return fallbackStatus, codes.NotFound, publicErrorMessage(err, fallbackStatus)
	case http.StatusConflict:
		return fallbackStatus, codes.Aborted, publicErrorMessage(err, fallbackStatus)
	case http.StatusPreconditionFailed:
		return fallbackStatus, codes.FailedPrecondition, publicErrorMessage(err, fallbackStatus)
	case http.StatusMethodNotAllowed:
		return fallbackStatus, codes.Unimplemented, publicErrorMessage(err, fallbackStatus)
	case http.StatusBadGateway:
		return fallbackStatus, codes.Unavailable, "upstream dependency failed"
	case http.StatusServiceUnavailable:
		return fallbackStatus, codes.Unavailable, "service unavailable"
	default:
		return http.StatusInternalServerError, codes.Internal, "internal server error"
	}
}

func publicErrorMessage(err error, status int) string {
	if err != nil && strings.TrimSpace(err.Error()) != "" {
		return err.Error()
	}
	return http.StatusText(status)
}
