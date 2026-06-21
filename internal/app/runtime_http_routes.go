package app

import (
	"net/http"

	httpapi "github.com/byte-v-forge/proxy-gateway/internal/app/httpapi"
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

func (api *runtimeHTTPAPI) authHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/auth/ws-token", Handler: api.handleAuthWebSocketToken},
	}
}

func (api *runtimeHTTPAPI) checkHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodPost}, Path: "/proxy_exit_ip", Handler: api.handleGetProxyExitIP},
		{Methods: []string{http.MethodPost}, Path: "/proxy_exit_geo", Handler: api.handleGetProxyExitGeo},
		{Methods: []string{http.MethodPost}, Path: "/ip_fraud_check", Handler: api.handleCheckIPFraud},
		{Methods: []string{http.MethodPost}, Path: "/check_cf_access_risk", Handler: api.handleCheckEdgeAccessRisk},
		{Methods: []string{http.MethodPost}, Path: "/target_connectivity_check", Handler: api.handleCheckTargetConnectivity},
		{Methods: []string{http.MethodGet, http.MethodPost}, Path: "/proxy_exit_check_snapshot", Handler: api.handleGetProxyExitCheckSnapshot},
	}
}

func (api *runtimeHTTPAPI) leaseHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/leases", Handler: api.handleLeases},
		{Methods: []string{http.MethodGet}, Path: "/leases/:lease_id", Handler: api.handleLease},
		{Methods: []string{http.MethodPost}, Path: "/leases/acquire", Handler: api.handleAcquireLease},
		{Methods: []string{http.MethodPost}, Path: "/leases/release", Handler: api.handleReleaseLease},
	}
}

func (api *runtimeHTTPAPI) providerHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/providers", Handler: api.handleProviders},
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, Path: "/provider-accounts", Handler: api.handleProviderAccounts},
	}
}

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

func (api *runtimeHTTPAPI) settingsHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, Path: "/settings/dynamic-ip-providers", Handler: api.handleDynamicIPProviders},
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, Path: "/settings/in-user-rules", Handler: api.handleInUserRules},
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, Path: "/settings/mihomo-native", Handler: api.handleMihomoNativeConfig},
		{Methods: []string{http.MethodGet}, Path: "/settings/ip-fraud-providers", Handler: api.handleIPFraudProviders},
		{Methods: []string{http.MethodGet}, Path: "/settings/ip-geo-providers", Handler: api.handleIPGeoProviders},
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, Path: "/settings", Handler: api.handleRuntimeSettings},
	}
}

func (api *runtimeHTTPAPI) runtimeStatusHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/runtime/status", Handler: api.handleRuntimeStatus},
		{Methods: []string{http.MethodGet}, Path: "/runtime/metrics/summary", Handler: api.handleRuntimeMetricsSummary},
	}
}
