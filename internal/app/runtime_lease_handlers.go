package app

import (
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleLeases(ctx *gin.Context) {
	options, err := parseLeaseListOptions(ctx)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	response, err := api.service.listProxyDynamicLeases(ctx.Request.Context(), options)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleLease(ctx *gin.Context) {
	leaseID := strings.TrimSpace(ctx.Param("lease_id"))
	lease, err := api.service.getProxyDynamicLease(ctx.Request.Context(), leaseID)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, lease)
}

func (api *runtimeHTTPAPI) handleAcquireLease(ctx *gin.Context) {
	var body proxyruntimev1.AcquireProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.service.acquireProxyLease(ctx.Request.Context(), ctx.Request, &body)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func parseLeaseListOptions(ctx *gin.Context) (leaseapp.ListOptions, error) {
	options, err := leaseapp.ParseListOptions(ctx.Request.URL.Query())
	if err != nil {
		return leaseapp.ListOptions{}, invalidArgument(err.Error(), err)
	}
	return options, nil
}

func (api *runtimeHTTPAPI) handleReleaseLease(ctx *gin.Context) {
	var body proxyruntimev1.ReleaseProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.service.ReleaseProxyLease(ctx.Request.Context(), &body)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}
