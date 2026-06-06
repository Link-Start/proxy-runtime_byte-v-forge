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
