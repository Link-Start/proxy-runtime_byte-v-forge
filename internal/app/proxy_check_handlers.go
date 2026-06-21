package app

import (
	"net/http"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleGetProxyExitIP(ctx *gin.Context) {
	var checkReq proxygatewayv1.GetProxyExitIPRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.checks.GetProxyExitIP(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleGetProxyExitGeo(ctx *gin.Context) {
	var checkReq proxygatewayv1.GetProxyExitGeoRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.checks.GetProxyExitGeo(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckIPFraud(ctx *gin.Context) {
	var checkReq proxygatewayv1.CheckProxyIPFraudRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.checks.CheckProxyIPFraud(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckEdgeAccessRisk(ctx *gin.Context) {
	var checkReq proxygatewayv1.CheckProxyEdgeAccessRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.checks.CheckProxyEdgeAccess(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleCheckTargetConnectivity(ctx *gin.Context) {
	var checkReq proxygatewayv1.CheckProxyTargetConnectivityRequest
	if !api.readOptionalProto(ctx, &checkReq) {
		return
	}
	response, err := api.checks.CheckProxyTargetConnectivity(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleGetProxyExitCheckSnapshot(ctx *gin.Context) {
	var checkReq proxygatewayv1.GetProxyExitCheckSnapshotRequest
	switch ctx.Request.Method {
	case http.MethodGet:
		checkReq.ListenerId = ctx.Query("listener_id")
	case http.MethodPost:
		if !api.readOptionalProto(ctx, &checkReq) {
			return
		}
	}
	response, err := api.checks.GetProxyExitCheckSnapshot(ctx.Request.Context(), &checkReq)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}
