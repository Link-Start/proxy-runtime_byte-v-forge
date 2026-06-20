package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func (api *runtimeHTTPAPI) handleHealth(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleReady(ctx *gin.Context) {
	if api.ready != nil {
		ready, msg := api.ready()
		if ready {
			ctx.Status(http.StatusNoContent)
			return
		}
		msg = appcore.FirstNonEmpty(msg, "route runtime is not running")
		writeHTTPError(ctx.Writer, appcore.Unavailable(msg, nil), http.StatusServiceUnavailable)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleRuntimeStatus(ctx *gin.Context) {
	response, err := api.status.GetProxyRuntimeStatus(ctx.Request.Context())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}

func (api *runtimeHTTPAPI) handleProviders(ctx *gin.Context) {
	response, err := api.providers.ListProxyProviders(ctx.Request.Context())
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}
