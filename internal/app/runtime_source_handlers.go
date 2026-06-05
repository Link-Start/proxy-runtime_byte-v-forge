package app

import (
	"context"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleSources(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		response, err := api.service.ListProxySources(req.Context(), &proxyruntimev1.ListProxySourcesRequest{})
		if err != nil {
			writeHTTPError(w, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(w, response)
	case http.MethodPost, http.MethodPut:
		var body proxyruntimev1.UpsertProxySubscriptionSourceRequest
		if !api.readProto(w, req, &body) {
			return
		}
		response, err := api.service.UpsertProxySubscriptionSource(req.Context(), &body)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	case http.MethodDelete:
		var body proxyruntimev1.DeleteProxySourceRequest
		if !api.readProto(w, req, &body) {
			return
		}
		response, err := api.service.DeleteProxySource(req.Context(), &body)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}

func (api *runtimeHTTPAPI) handleFixedSources(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost, http.MethodPut:
		var body proxyruntimev1.UpsertProxyFixedSourceRequest
		if !api.readProto(w, req, &body) {
			return
		}
		response, err := api.service.UpsertProxyFixedSource(req.Context(), &body)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodPost+", "+http.MethodPut)
	}
}

func (api *runtimeHTTPAPI) handleSourceNodes(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	response, err := api.service.ListProxySourceNodes(req.Context(), &proxyruntimev1.ListProxySourceNodesRequest{SourceId: req.URL.Query().Get("source_id")})
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (r *Runtime) listSources(ctx context.Context) ([]*proxyruntimev1.ProxySourceDescriptor, error) {
	settings, err := r.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	sources, err := r.store.ListSources(ctx, dynamicIPGatewayMap(settings))
	if err != nil {
		return nil, err
	}
	return sources, nil
}
