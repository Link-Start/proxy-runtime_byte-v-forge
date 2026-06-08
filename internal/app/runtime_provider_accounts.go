package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleProviderAccounts(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		response, err := api.service.ListProxyProviderAccounts(ctx.Request.Context(), &proxyruntimev1.ListProxyProviderAccountsRequest{})
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(ctx, response)
	case http.MethodPost, http.MethodPut:
		var body proxyruntimev1.UpsertProxyProviderAccountRequest
		if !api.readProto(ctx, &body) {
			return
		}
		response, err := api.service.UpsertProxyProviderAccount(ctx.Request.Context(), &body)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	case http.MethodDelete:
		var body proxyruntimev1.DeleteProxyProviderAccountRequest
		if !api.readProto(ctx, &body) {
			return
		}
		response, err := api.service.DeleteProxyProviderAccount(ctx.Request.Context(), &body)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	}
}
