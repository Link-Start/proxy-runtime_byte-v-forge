package app

import (
	"net/http"
	"strings"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) registerMihomoDashboardRoutes(router *gin.Engine) {
	router.GET("/mihomo/dashboard", api.handleMihomoDashboard)
	router.GET("/mihomo/ui", redirectToTrailingSlash)
	router.Any("/mihomo/ui/*path", api.mihomoReverseProxy("/mihomo/ui/", "/ui/"))
	router.Any("/mihomo/controller", api.mihomoReverseProxy("/mihomo/controller", "/"))
	router.Any("/mihomo/controller/*path", api.mihomoReverseProxy("/mihomo/controller/", "/"))
}

func (api *runtimeHTTPAPI) handleMihomoDashboard(ctx *gin.Context) {
	api.writeMihomoDashboardBootstrap(ctx, "/mihomo/controller", "/mihomo/ui/#/overview")
}

func (api *runtimeHTTPAPI) writeMihomoDashboardBootstrap(ctx *gin.Context, endpointURL string, uiURL string) {
	body, err := dashboardapp.BootstrapHTML(dashboardapp.BootstrapOptions{
		EndpointURL:  endpointURL,
		UIURL:        uiURL,
		AuthRequired: strings.TrimSpace(api.authToken) != "",
	})
	if err != nil {
		writeHTTPError(ctx.Writer, err, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Cache-Control", "no-store")
	_, _ = ctx.Writer.Write(body)
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

func redirectToTrailingSlash(ctx *gin.Context) {
	target := *ctx.Request.URL
	target.Path += "/"
	ctx.Redirect(http.StatusTemporaryRedirect, target.String())
}
