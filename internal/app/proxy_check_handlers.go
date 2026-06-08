package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleGetProxyExitIP(ctx *gin.Context) {
	var checkReq proxyruntimev1.GetProxyExitIPRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.service.GetProxyExitIP(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleGetProxyExitGeo(ctx *gin.Context) {
	var checkReq proxyruntimev1.GetProxyExitGeoRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.service.GetProxyExitGeo(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckIPFraud(ctx *gin.Context) {
	var checkReq proxyruntimev1.CheckProxyIPFraudRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyIPFraud(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckEdgeAccessRisk(ctx *gin.Context) {
	var checkReq proxyruntimev1.CheckProxyEdgeAccessRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyEdgeAccess(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckTargetConnectivity(ctx *gin.Context) {
	var checkReq proxyruntimev1.CheckProxyTargetConnectivityRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.service.CheckProxyTargetConnectivity(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleGetProxyExitCheckSnapshot(ctx *gin.Context) {
	var checkReq proxyruntimev1.GetProxyExitCheckSnapshotRequest
	switch ctx.Request.Method {
	case http.MethodGet:
		checkReq.ListenerId = ctx.Query("listener_id")
	case http.MethodPost:
		if !api.readOptionalProto(ctx, &checkReq) {
			return
		}
	}
	response, err := api.service.GetProxyExitCheckSnapshot(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleIPFraudProviders(ctx *gin.Context) {
	response, err := api.service.ListProxyIPFraudProviders(ctx.Request.Context(), &proxyruntimev1.ListProxyIPFraudProvidersRequest{})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleIPGeoProviders(ctx *gin.Context) {
	response, err := api.service.ListProxyIPGeoProviders(ctx.Request.Context(), &proxyruntimev1.ListProxyIPGeoProvidersRequest{})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleDynamicIPProviders(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		response, err := api.service.GetProxyRuntimeSettings(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeSettingsRequest{})
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(ctx, response)
	case http.MethodPost, http.MethodPut:
		var updateReq proxyruntimev1.UpdateProxyRuntimeSettingsRequest
		if !api.readProto(ctx, &updateReq) {
			return
		}
		response, err := api.service.UpdateProxyDynamicIPProviders(ctx.Request.Context(), &updateReq)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	}
}

func (api *runtimeHTTPAPI) handleRuntimeSettings(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		response, err := api.service.GetProxyRuntimeSettings(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeSettingsRequest{})
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(ctx, response)
	case http.MethodPost, http.MethodPut:
		var updateReq proxyruntimev1.UpdateProxyRuntimeSettingsRequest
		if !api.readProto(ctx, &updateReq) {
			return
		}
		response, err := api.service.UpdateProxyRuntimeSettings(ctx.Request.Context(), &updateReq)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	}
}
