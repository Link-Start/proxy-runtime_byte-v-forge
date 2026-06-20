package app

import (
	"errors"
	"net/http"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func writeLeaseHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	if errors.Is(err, leaseapp.ErrLeaseIDRequired) {
		writeHTTPError(w, appcore.InvalidArgument(err.Error(), err), http.StatusBadRequest)
		return
	}
	if isStoreNotFound(err) {
		writeHTTPError(w, errors.New("lease not found"), http.StatusNotFound)
		return
	}
	writeHTTPError(w, err, fallbackStatus)
}
