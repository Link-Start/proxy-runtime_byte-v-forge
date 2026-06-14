package app

import (
	"net/http"
	"strconv"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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

func parseLeaseListOptions(ctx *gin.Context) (leaseListOptions, error) {
	options := defaultLeaseListOptions()
	status := strings.TrimSpace(ctx.Query("status"))
	if status == "" && ctx.Query("include_inactive") == "true" {
		status = string(leaseListModeRecent)
	}
	if status != "" {
		switch leaseListMode(status) {
		case leaseListModeActive, leaseListModeRecent, leaseListModeHistory:
			options.mode = leaseListMode(status)
		default:
			return leaseListOptions{}, invalidArgument("unsupported lease list status", nil)
		}
	}
	if rawLimit := strings.TrimSpace(ctx.Query("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return leaseListOptions{}, invalidArgument("lease list limit must be a positive integer", err)
		}
		options.limit = limit
	}
	return normalizeLeaseListOptions(options), nil
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
