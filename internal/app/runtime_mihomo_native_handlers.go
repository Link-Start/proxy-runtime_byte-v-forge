package app

import (
	"errors"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	settingsapp "github.com/byte-v-forge/proxy-runtime/internal/app/settings/application"
	"github.com/gin-gonic/gin"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
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
	response, err := api.settings.GetMihomoNative(ctx.Request.Context(), &proxyruntimev1.GetProxyRuntimeMihomoNativeConfigRequest{})
	if err != nil {
		writeSettingsLoadHTTPError(ctx, appcore.InternalError("load mihomo native config", err))
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleUpdateMihomoNativeConfig(ctx *gin.Context) {
	var req proxyruntimev1.UpdateProxyRuntimeMihomoNativeConfigRequest
	if !api.readProto(ctx, &req) {
		return
	}
	response, err := api.settings.UpdateMihomoNative(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, settingsapp.ErrMihomoNativeUpdateUnavailable) {
			err = appcore.InternalError("mihomo native settings update unavailable", err)
		}
		writeSettingsUpdateHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}
