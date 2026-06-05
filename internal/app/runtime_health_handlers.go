package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleReady(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if api.ready != nil {
		ready, msg := api.ready()
		if ready {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		msg = firstNonEmpty(msg, "route runtime is not running")
		writeHTTPError(w, unavailable(msg, nil), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleProviders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	response, err := api.service.ListProxyProviders(req.Context(), &proxyruntimev1.ListProxyProvidersRequest{})
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleGateway(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	response, err := api.service.GetEgressGateway(req.Context(), &proxyruntimev1.GetEgressGatewayRequest{})
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handlePool(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	response, err := api.service.GetProxyPool(req.Context(), &proxyruntimev1.GetProxyPoolRequest{})
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	response, err := api.service.RefreshProxyPool(req.Context(), &proxyruntimev1.RefreshProxyPoolRequest{})
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}
