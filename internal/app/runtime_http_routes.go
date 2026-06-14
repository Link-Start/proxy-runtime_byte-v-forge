package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
)

const controlPlaneHTTPPrefix = "/api"

func (api *runtimeHTTPAPI) registerControlPlaneHTTPRoutes(router *gin.Engine) {
	group := router.Group(controlPlaneHTTPPrefix)
	for _, route := range api.controlPlaneHTTPRoutes() {
		group.Match(route.Methods, route.Path, route.Handler)
	}
	api.registerMihomoDashboardRoutes(router)
}

func (api *runtimeHTTPAPI) controlPlaneHTTPRoutes() []httpapi.Route {
	routes := []httpapi.Route{}
	routes = append(routes, api.authHTTPRoutes()...)
	routes = append(routes, api.runtimeStatusHTTPRoutes()...)
	routes = append(routes, api.providerHTTPRoutes()...)
	routes = append(routes, api.leaseHTTPRoutes()...)
	routes = append(routes, api.checkHTTPRoutes()...)
	routes = append(routes, api.settingsHTTPRoutes()...)
	return routes
}
