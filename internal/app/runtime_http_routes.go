package app

import "net/http"

type runtimeHTTPRoute struct {
	path    string
	handler http.HandlerFunc
}

var controlPlaneHTTPPrefixes = []string{"/proxy", "/api/proxy-runtime"}

func (api *runtimeHTTPAPI) registerControlPlaneHTTPRoutes(mux *http.ServeMux) {
	for _, prefix := range controlPlaneHTTPPrefixes {
		for _, route := range api.controlPlaneHTTPRoutes() {
			mux.HandleFunc(prefix+route.path, route.handler)
		}
		api.registerMihomoDashboardRoutes(mux, prefix)
	}
}

func (api *runtimeHTTPAPI) controlPlaneHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{path: "/providers", handler: api.handleProviders},
		{path: "/provider-accounts", handler: api.handleProviderAccounts},
		{path: "/leases", handler: api.handleLeases},
		{path: "/leases/acquire", handler: api.handleAcquireLease},
		{path: "/leases/release", handler: api.handleReleaseLease},
		{path: "/proxy_exit_ip", handler: api.handleGetProxyExitIP},
		{path: "/proxy_exit_geo", handler: api.handleGetProxyExitGeo},
		{path: "/ip_fraud_check", handler: api.handleCheckIPFraud},
		{path: "/check_cf_access_risk", handler: api.handleCheckEdgeAccessRisk},
		{path: "/target_connectivity_check", handler: api.handleCheckTargetConnectivity},
		{path: "/settings/dynamic-ip-providers", handler: api.handleDynamicIPProviders},
		{path: "/settings/in-user-rules", handler: api.handleInUserRules},
		{path: "/settings/mihomo-native", handler: api.handleMihomoNativeConfig},
		{path: "/settings/ip-fraud-providers", handler: api.handleIPFraudProviders},
		{path: "/settings/ip-geo-providers", handler: api.handleIPGeoProviders},
		{path: "/settings", handler: api.handleRuntimeSettings},
	}
}
