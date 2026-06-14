package app

import (
	"net/http"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) registerMihomoDashboardRoutes(router *gin.Engine) {
	router.GET("/mihomo/dashboard", api.handleMihomoDashboard)
	router.GET("/mihomo/ui", httpapi.RedirectToTrailingSlash)
	router.Any("/mihomo/ui/*path", api.mihomoReverseProxy("/mihomo/ui/", "/ui/"))
	router.Any("/mihomo/controller", api.mihomoReverseProxy("/mihomo/controller", "/"))
	router.Any("/mihomo/controller/*path", api.mihomoReverseProxy("/mihomo/controller/", "/"))
}

func (api *runtimeHTTPAPI) handleMihomoDashboard(ctx *gin.Context) {
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/overview")
}

func (api *runtimeHTTPAPI) writeMihomoDashboardBootstrap(ctx *gin.Context, endpointURL string, uiURL string) {
	err := dashboardapp.WriteBootstrap(ctx.Writer, dashboardapp.BootstrapOptions{
		EndpointURL:  endpointURL,
		UIURL:        uiURL,
		AuthRequired: api.auth.Enabled(),
	})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
	}
}

func (api *runtimeHTTPAPI) mihomoReverseProxy(mountPrefix string, upstreamPrefix string) gin.HandlerFunc {
	proxy, err := dashboardapp.NewReverseProxy(dashboardapp.ReverseProxyOptions{
		APIAddr:        api.mihomoAPIAddr,
		MountPrefix:    mountPrefix,
		UpstreamPrefix: upstreamPrefix,
		AuthToken:      api.authToken,
		WriteError:     writeHTTPError,
	})
	if err != nil {
		return func(ctx *gin.Context) {
			writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
		}
	}
	return gin.WrapH(proxy)
}
