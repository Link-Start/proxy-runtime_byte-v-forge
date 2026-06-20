package app

import (
	"context"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

type runtimeSettingsUpdateHandler func(
	context.Context,
	*proxyruntimev1.UpdateProxyRuntimeSettingsRequest,
) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error)

func (api *runtimeHTTPAPI) handleRuntimeSettings(ctx *gin.Context) {
	api.handleRuntimeSettingsViewOrUpdate(ctx, api.settings.UpdateRuntimeSettings)
}

func (api *runtimeHTTPAPI) handleDynamicIPProviders(ctx *gin.Context) {
	api.handleRuntimeSettingsViewOrUpdate(ctx, func(reqCtx context.Context, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest) (*proxyruntimev1.UpdateProxyRuntimeSettingsResponse, error) {
		return api.settings.UpdateDynamicIPProviders(reqCtx, req.GetDynamicIpProviders())
	})
}

func (api *runtimeHTTPAPI) handleInUserRules(ctx *gin.Context) {
	api.handleRuntimeSettingsViewOrUpdate(ctx, api.settings.UpdateInUserRules)
}

func (api *runtimeHTTPAPI) handleRuntimeSettingsViewOrUpdate(ctx *gin.Context, update runtimeSettingsUpdateHandler) {
	switch ctx.Request.Method {
	case http.MethodGet:
		api.handleGetRuntimeSettings(ctx)
	case http.MethodPost, http.MethodPut:
		api.handleUpdateRuntimeSettings(ctx, update)
	}
}

func (api *runtimeHTTPAPI) handleGetRuntimeSettings(ctx *gin.Context) {
	response, err := api.settings.Get(ctx.Request.Context())
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleUpdateRuntimeSettings(ctx *gin.Context, update runtimeSettingsUpdateHandler) {
	var updateReq proxyruntimev1.UpdateProxyRuntimeSettingsRequest
	if !api.readProto(ctx, &updateReq) {
		return
	}
	response, err := update(ctx.Request.Context(), &updateReq)
	if err != nil {
		writeSettingsUpdateHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}
