package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleProviderAccounts(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		api.handleListProviderAccounts(ctx)
	case http.MethodPost, http.MethodPut:
		api.handleUpsertProviderAccount(ctx)
	case http.MethodDelete:
		api.handleDeleteProviderAccount(ctx)
	}
}

func (api *runtimeHTTPAPI) handleListProviderAccounts(ctx *gin.Context) {
	response, err := api.providers.ListProxyProviderAccounts(ctx.Request.Context())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleUpsertProviderAccount(ctx *gin.Context) {
	var body proxyruntimev1.UpsertProxyProviderAccountRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.providers.UpsertProxyProviderAccount(ctx.Request.Context(), &body)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleDeleteProviderAccount(ctx *gin.Context) {
	var body proxyruntimev1.DeleteProxyProviderAccountRequest
	if !api.readProto(ctx, &body) {
		return
	}
	response, err := api.providers.DeleteProxyProviderAccount(ctx.Request.Context(), &body)
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
		return
	}
	api.writeProto(ctx, response)
}
