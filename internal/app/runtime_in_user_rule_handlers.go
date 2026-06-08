package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleInUserRules(ctx *gin.Context) {
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
		response, err := api.service.UpdateProxyInUserRules(ctx.Request.Context(), &updateReq)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	}
}
