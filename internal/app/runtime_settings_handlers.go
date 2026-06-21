package app

import (
	"context"
	"net/http"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/gin-gonic/gin"
)

type runtimeSettingsUpdateHandler func(
	context.Context,
	*proxygatewayv1.UpdateProxyGatewaySettingsRequest,
) (*proxygatewayv1.UpdateProxyGatewaySettingsResponse, error)

func (api *runtimeHTTPAPI) handleRuntimeSettings(ctx *gin.Context) {
	api.handleRuntimeSettingsViewOrUpdate(ctx, api.settings.UpdateRuntimeSettings)
}

func (api *runtimeHTTPAPI) handleDynamicIPProviders(ctx *gin.Context) {
	api.handleRuntimeSettingsViewOrUpdate(ctx, func(reqCtx context.Context, req *proxygatewayv1.UpdateProxyGatewaySettingsRequest) (*proxygatewayv1.UpdateProxyGatewaySettingsResponse, error) {
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
	var updateReq proxygatewayv1.UpdateProxyGatewaySettingsRequest
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

func (api *runtimeHTTPAPI) handleIPFraudProviders(ctx *gin.Context) {
	response, err := api.settings.ListIPFraudProviders(ctx.Request.Context())
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleIPGeoProviders(ctx *gin.Context) {
	response, err := api.settings.ListIPGeoProviders(ctx.Request.Context())
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func writeSettingsLoadHTTPError(ctx *gin.Context, err error) {
	writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
}

func writeSettingsUpdateHTTPError(ctx *gin.Context, err error) {
	writeHTTPError(ctx.Writer, err, http.StatusBadRequest)
}
