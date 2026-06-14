package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet}, path: "/auth/ws-token", handler: api.handleAuthWebSocketToken},
		{methods: []string{http.MethodGet}, path: "/runtime/status", handler: api.handleRuntimeStatus},
		{methods: []string{http.MethodGet}, path: "/providers", handler: api.handleProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, path: "/provider-accounts", handler: api.handleProviderAccounts},
		{methods: []string{http.MethodGet}, path: "/leases", handler: api.handleLeases},
		{methods: []string{http.MethodPost}, path: "/leases/acquire", handler: api.handleAcquireLease},
		{methods: []string{http.MethodPost}, path: "/leases/release", handler: api.handleReleaseLease},
		{methods: []string{http.MethodPost}, path: "/proxy_exit_ip", handler: api.handleGetProxyExitIP},
		{methods: []string{http.MethodPost}, path: "/proxy_exit_geo", handler: api.handleGetProxyExitGeo},
		{methods: []string{http.MethodPost}, path: "/ip_fraud_check", handler: api.handleCheckIPFraud},
		{methods: []string{http.MethodPost}, path: "/check_cf_access_risk", handler: api.handleCheckEdgeAccessRisk},
		{methods: []string{http.MethodPost}, path: "/target_connectivity_check", handler: api.handleCheckTargetConnectivity},
		{methods: []string{http.MethodGet, http.MethodPost}, path: "/proxy_exit_check_snapshot", handler: api.handleGetProxyExitCheckSnapshot},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/dynamic-ip-providers", handler: api.handleDynamicIPProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/in-user-rules", handler: api.handleInUserRules},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/mihomo-native", handler: api.handleMihomoNativeConfig},
		{methods: []string{http.MethodGet}, path: "/settings/ip-fraud-providers", handler: api.handleIPFraudProviders},
		{methods: []string{http.MethodGet}, path: "/settings/ip-geo-providers", handler: api.handleIPGeoProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings", handler: api.handleRuntimeSettings},
	}
}
