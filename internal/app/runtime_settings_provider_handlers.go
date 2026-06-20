package app

import (
	"github.com/gin-gonic/gin"
)

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
