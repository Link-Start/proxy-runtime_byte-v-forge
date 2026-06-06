package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (api *runtimeHTTPAPI) handleGetProxyExitIP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var checkReq proxyruntimev1.GetProxyExitIPRequest
	if !api.readOptionalProto(w, req, &checkReq) {
		return
	}
	response, err := api.service.GetProxyExitIP(req.Context(), &checkReq)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleGetProxyExitGeo(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var checkReq proxyruntimev1.GetProxyExitGeoRequest
	if !api.readOptionalProto(w, req, &checkReq) {
		return
	}
	response, err := api.service.GetProxyExitGeo(req.Context(), &checkReq)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleCheckIPFraud(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var checkReq proxyruntimev1.CheckProxyIPFraudRequest
	if !api.readOptionalProto(w, req, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyIPFraud(req.Context(), &checkReq)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleCheckEdgeAccessRisk(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var checkReq proxyruntimev1.CheckProxyEdgeAccessRequest
	if !api.readOptionalProto(w, req, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyEdgeAccess(req.Context(), &checkReq)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleCheckTargetConnectivity(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var checkReq proxyruntimev1.CheckProxyTargetConnectivityRequest
	if !api.readOptionalProto(w, req, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyTargetConnectivity(req.Context(), &checkReq)
	if err != nil {
		writeHTTPError(w, err, http.StatusBadGateway)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleIPFraudProviders(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	response, err := api.service.ListProxyIPFraudProviders(req.Context(), &proxyruntimev1.ListProxyIPFraudProvidersRequest{})
	if err != nil {
		writeHTTPError(w, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(w, response)
}

func (api *runtimeHTTPAPI) handleDynamicIPProviders(w http.ResponseWriter, req *http.Request) {
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
		response, err := api.service.UpdateProxyDynamicIPProviders(req.Context(), &updateReq)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut)
	}
}

func (api *runtimeHTTPAPI) handleRuntimeSettings(w http.ResponseWriter, req *http.Request) {
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
		response, err := api.service.UpdateProxyRuntimeSettings(req.Context(), &updateReq)
		if err != nil {
			writeHTTPError(w, err, http.StatusBadRequest)
			return
		}
		api.writeProto(w, response)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut)
	}
}
