package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleMihomoNativeConfig(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		api.handleGetMihomoNativeConfig(ctx)
	case http.MethodPost, http.MethodPut:
		api.handleUpdateMihomoNativeConfig(ctx)
	}
}

func (api *runtimeHTTPAPI) handleGetMihomoNativeConfig(ctx *gin.Context) {
	response, err := api.service.GetProxyRuntimeMihomoNativeConfig(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest{})
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleUpdateMihomoNativeConfig(ctx *gin.Context) {
	var req proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest
	if !api.readProto(ctx, &req) {
		return
	}
	response, err := api.service.UpdateProxyRuntimeMihomoNativeConfig(ctx.Request.Context(), &req)
	if err != nil {
		writeSettingsUpdateHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}
