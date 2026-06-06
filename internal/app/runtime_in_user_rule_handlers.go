package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleInUserRules(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		response, err := api.service.GetProxyRuntimeSettings(req.Context(), &proxyruntimev1.GetProxyRuntimeSettingsRequest{})
		if err != nil {
			writeHTTPError(w, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(w, response)
	case http.MethodPost, http.MethodPut:
		var updateReq proxyruntimev1.UpdateProxyRuntimeSettingsRequest
		if !api.readProto(w, req, &updateReq) {
			return
		}
		response, err := api.service.UpdateProxyInUserRules(req.Context(), &updateReq)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut)
	}
}
