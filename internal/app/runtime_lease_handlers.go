package app

import (
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/gin-gonic/gin"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (api *runtimeHTTPAPI) handleLeases(ctx *gin.Context) {
	options, err := parseLeaseListOptions(ctx)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	startedAt := time.Now()
	response, err := api.leases.ListProxyDynamicLeaseFacts(ctx.Request.Context(), options)
	api.observe(runtimeMetricLeaseList, startedAt, err)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleLease(ctx *gin.Context) {
	leaseID := strings.TrimSpace(ctx.Param("lease_id"))
	lease, err := api.leases.GetProxyDynamicLeaseFact(ctx.Request.Context(), leaseID)
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
	startedAt := time.Now()
	response, err := api.leases.AcquireProxyLease(ctx.Request.Context(), advertisedProxyHost(ctx.Request), &body)
	api.observe(runtimeMetricLeaseAcquire, startedAt, err)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}

func parseLeaseListOptions(ctx *gin.Context) (leaseapp.ListOptions, error) {
	options, err := leaseapp.ParseListOptions(ctx.Request.URL.Query())
	if err != nil {
		return leaseapp.ListOptions{}, appcore.InvalidArgument(err.Error(), err)
	}
	return options, nil
}

func (api *runtimeHTTPAPI) handleReleaseLease(ctx *gin.Context) {
	var body proxyruntimev1.ReleaseProxyLeaseRequest
	if !api.readProto(ctx, &body) {
		return
	}
	startedAt := time.Now()
	response, err := api.leases.ReleaseProxyLease(ctx.Request.Context(), &body)
	api.observe(runtimeMetricLeaseRelease, startedAt, err)
	if err != nil {
		writeLeaseHTTPError(ctx.Writer, err, http.StatusBadGateway)
		return
	}
	api.writeProto(ctx, response)
}
