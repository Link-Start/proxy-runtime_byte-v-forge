package app

import "github.com/gin-gonic/gin"

type runtimeHTTPRoute struct {
	methods []string
	path    string
	handler gin.HandlerFunc
}

const controlPlaneHTTPPrefix = "/api"

func (api *runtimeHTTPAPI) registerControlPlaneHTTPRoutes(router *gin.Engine) {
	group := router.Group(controlPlaneHTTPPrefix)
	for _, route := range api.controlPlaneHTTPRoutes() {
		group.Match(route.methods, route.path, route.handler)
	}
	api.registerMihomoDashboardRoutes(router)
}

func (api *runtimeHTTPAPI) controlPlaneHTTPRoutes() []runtimeHTTPRoute {
	routes := []runtimeHTTPRoute{}
	routes = append(routes, api.authHTTPRoutes()...)
	routes = append(routes, api.runtimeStatusHTTPRoutes()...)
	routes = append(routes, api.providerHTTPRoutes()...)
	routes = append(routes, api.leaseHTTPRoutes()...)
	routes = append(routes, api.checkHTTPRoutes()...)
	routes = append(routes, api.settingsHTTPRoutes()...)
	return routes
}
