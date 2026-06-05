package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleResolveProxy(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var body proxyruntimev1.ResolveProxyRequest
	if !api.readProto(w, req, &body) {
		return
	}
	response, err := api.service.resolveProxy(req.Context(), req, &body)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}
