package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) handleIPFraudProviders(ctx *gin.Context) {
	response, err := api.service.ListProxyIPFraudProviders(ctx.Request.Context(), &proxyruntimev1.ListProxyIPFraudProvidersRequest{})
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleIPGeoProviders(ctx *gin.Context) {
	response, err := api.service.ListProxyIPGeoProviders(ctx.Request.Context(), &proxyruntimev1.ListProxyIPGeoProvidersRequest{})
	if err != nil {
		writeSettingsLoadHTTPError(ctx, err)
		return
	}
	api.writeProto(ctx, response)
}
