package app

import (
	"net/http"

	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) registerPublicHTTPRoutes(router *gin.Engine) {
	for _, route := range api.publicHTTPRoutes() {
		router.Match(route.Methods, route.Path, route.Handler)
	}
}

func (api *runtimeHTTPAPI) publicHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/healthz", Handler: api.handleHealth},
		{Methods: []string{http.MethodGet}, Path: "/readyz", Handler: api.handleReady},
		{Methods: []string{http.MethodGet}, Path: "/metrics", Handler: api.handleMetrics},
		{Methods: []string{http.MethodGet}, Path: "/login", Handler: api.handleAuthLoginPage},
		{Methods: []string{http.MethodGet}, Path: "/api/auth/session", Handler: api.handleAuthSession},
		{Methods: []string{http.MethodPost}, Path: "/api/auth/login", Handler: api.handleAuthLogin},
		{Methods: []string{http.MethodPost}, Path: "/api/auth/logout", Handler: api.handleAuthLogout},
		{Methods: []string{http.MethodGet, http.MethodHead}, Path: "/", Handler: api.handleDashboardEntry},
	}
}
