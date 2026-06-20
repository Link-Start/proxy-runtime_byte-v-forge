// Package appcore holds cross-cutting leaf helpers (structured errors, id and
// string utilities) shared by the app package and its subpackages. It depends
// on nothing inside internal/app so any app subpackage can import it without a
// cycle.
package appcore

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/grpc/codes"
)

type errorCode string

const (
	codeInvalidArgument    errorCode = "invalid_argument"
	codeNotFound           errorCode = "not_found"
	codeFailedPrecondition errorCode = "failed_precondition"
	codeResourceExhausted  errorCode = "resource_exhausted"
	codeUnavailable        errorCode = "unavailable"
	codeInternal           errorCode = "internal"
)

type appError struct {
	code    errorCode
	message string
	cause   error
}

func (e *appError) Error() string { return e.message }

func (e *appError) Unwrap() error { return e.cause }

func newAppError(code errorCode, message string, cause error) error {
	message = strings.TrimSpace(message)
	if message == "" && cause != nil {
		message = cause.Error()
	}
	if message == "" {
		message = "request failed"
	}
	return &appError{code: code, message: message, cause: cause}
}

// InvalidArgument builds a structured invalid-argument error.
func InvalidArgument(message string, cause error) error {
	return newAppError(codeInvalidArgument, message, cause)
}

// FailedPrecondition builds a structured failed-precondition error.
func FailedPrecondition(message string, cause error) error {
	return newAppError(codeFailedPrecondition, message, cause)
}

// ResourceExhausted builds a structured resource-exhausted error.
func ResourceExhausted(message string, cause error) error {
	return newAppError(codeResourceExhausted, message, cause)
}

// Unavailable builds a structured unavailable error.
func Unavailable(message string, cause error) error {
	return newAppError(codeUnavailable, message, cause)
}

// InternalError builds a structured internal error.
func InternalError(message string, cause error) error {
	return newAppError(codeInternal, message, cause)
}

// IsAppError reports whether err is (or wraps) a structured app error.
func IsAppError(err error) bool {
	var appErr *appError
	return errors.As(err, &appErr)
}

// IsUnavailable reports whether err is a structured unavailable error.
func IsUnavailable(err error) bool { return hasCode(err, codeUnavailable) }

// IsFailedPrecondition reports whether err is a structured failed-precondition error.
func IsFailedPrecondition(err error) bool { return hasCode(err, codeFailedPrecondition) }

func hasCode(err error, code errorCode) bool {
	var appErr *appError
	return errors.As(err, &appErr) && appErr.code == code
}

// HTTPErrorDetails maps err to an HTTP status, gRPC code and a public message,
// using fallbackStatus when err is not a structured app error.
func HTTPErrorDetails(err error, fallbackStatus int) (int, codes.Code, string) {
	var appErr *appError
	if errors.As(err, &appErr) {
		switch appErr.code {
		case codeInvalidArgument:
			return http.StatusBadRequest, codes.InvalidArgument, appErr.message
		case codeNotFound:
			return http.StatusNotFound, codes.NotFound, appErr.message
		case codeFailedPrecondition:
			return http.StatusPreconditionFailed, codes.FailedPrecondition, appErr.message
		case codeResourceExhausted:
			return http.StatusRequestEntityTooLarge, codes.ResourceExhausted, appErr.message
		case codeUnavailable:
			return http.StatusBadGateway, codes.Unavailable, appErr.message
		case codeInternal:
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

// ErrorLogType returns the Go type name of err for operator-only logging. It
// never returns the error message, which may embed sensitive material.
func ErrorLogType(err error) string {
	if err == nil {
		return ""
	}
	return reflect.TypeOf(err).String()
}
