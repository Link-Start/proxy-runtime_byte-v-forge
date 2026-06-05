package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleLeases(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	includeInactive := req.URL.Query().Get("include_inactive") == "true"
	response, err := api.service.listProxyDynamicLeases(req.Context(), includeInactive)
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleAcquireLease(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var body proxyruntimev1.AcquireProxyLeaseRequest
	if !api.readProto(w, req, &body) {
		return
	}
	response, err := api.service.acquireProxyLease(req.Context(), req, &body)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleReleaseLease(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var body proxyruntimev1.ReleaseProxyLeaseRequest
	if !api.readProto(w, req, &body) {
		return
	}
	response, err := api.service.ReleaseProxyLease(req.Context(), &body)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}
