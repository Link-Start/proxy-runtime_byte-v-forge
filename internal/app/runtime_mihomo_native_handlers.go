package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleMihomoNativeConfig(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		response, err := api.service.GetProxyRuntimeMihomoNativeConfig(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest{})
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
			return
		}
		api.writeProto(ctx, response)
	case http.MethodPost, http.MethodPut:
		var req proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest
		if !api.readProto(ctx, &req) {
			return
		}
		response, err := api.service.UpdateProxyRuntimeMihomoNativeConfig(ctx.Request.Context(), &req)
		if err != nil {
			writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
			return
		}
		api.writeProto(ctx, response)
	}
}
