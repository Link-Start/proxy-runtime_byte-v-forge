package app

import (
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/gin-gonic/gin"
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
		msg = firstNonEmpty(msg, "route runtime is not running")
		writeHTTPError(ctx.Writer, unavailable(msg, nil), http.StatusServiceUnavailable)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (api *runtimeHTTPAPI) handleProviders(ctx *gin.Context) {
	response, err := api.service.ListProxyProviders(ctx.Request.Context(), &proxyruntimev1.ListProxyProvidersRequest{})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	api.writeProto(ctx, response)
}
