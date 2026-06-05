package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleProviderAccounts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		response, err := api.service.ListProxyProviderAccounts(req.Context(), &proxyruntimev1.ListProxyProviderAccountsRequest{})
		if err != nil {
			writeHTTPError(w, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(w, response)
	case http.MethodPost, http.MethodPut:
		var body proxyruntimev1.UpsertProxyProviderAccountRequest
		if !api.readProto(w, req, &body) {
			return
		}
		response, err := api.service.UpsertProxyProviderAccount(req.Context(), &body)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	case http.MethodDelete:
		var body proxyruntimev1.DeleteProxyProviderAccountRequest
		if !api.readProto(w, req, &body) {
			return
		}
		response, err := api.service.DeleteProxyProviderAccount(req.Context(), &body)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}
