package app

import (
	"errors"
	"net/http"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func writeLeaseHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	if errors.Is(err, leaseapp.ErrLeaseIDRequired) {
		writeHTTPError(w, invalidArgument(err.Error(), err), http.StatusBadRequest)
		return
	}
	if isStoreNotFound(err) {
		writeHTTPError(w, errors.New("lease not found"), http.StatusNotFound)
		return
	}
	writeHTTPError(w, err, fallbackStatus)
}
