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
	router.Any("/mihomo/ui/*path", api.dashboardProxies.ui)
	router.Any("/mihomo/controller", api.dashboardProxies.controller)
	router.Any("/mihomo/controller/*path", api.dashboardProxies.controller)
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

type dashboardProxyHandlers struct {
	fallback   gin.HandlerFunc
	ui         gin.HandlerFunc
	controller gin.HandlerFunc
}

func newDashboardProxyHandlers(apiAddr string, authToken string) dashboardProxyHandlers {
	bundle, err := dashboardapp.NewReverseProxyBundle(dashboardapp.ReverseProxyBundleOptions{
		APIAddr:    apiAddr,
		AuthToken:  authToken,
		WriteError: writeHTTPError,
	})
	if err != nil {
		errorHandler := dashboardProxyErrorHandler(err)
		return dashboardProxyHandlers{fallback: errorHandler, ui: errorHandler, controller: errorHandler}
	}
	return dashboardProxyHandlers{
		fallback:   gin.WrapH(bundle.Root),
		ui:         gin.WrapH(bundle.UI),
		controller: gin.WrapH(bundle.Controller),
	}
}

func dashboardProxyErrorHandler(err error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		writeHTTPError(ctx.Writer, err, http.StatusBadGateway)
	}
}
