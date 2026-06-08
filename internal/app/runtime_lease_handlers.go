package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleLeases(ctx *gin.Context) {
	includeInactive := ctx.Query("include_inactive") == "true"
	response, err := api.service.listProxyDynamicLeases(ctx.Request.Context(), includeInactive)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleAcquireLease(ctx *gin.Context) {
	var body proxyruntimev1.AcquireProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.service.acquireProxyLease(ctx.Request.Context(), ctx.Request, &body)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleReleaseLease(ctx *gin.Context) {
	var body proxyruntimev1.ReleaseProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.service.ReleaseProxyLease(ctx.Request.Context(), &body)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}
