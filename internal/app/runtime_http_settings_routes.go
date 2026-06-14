package app

import "net/http"

func (api *runtimeHTTPAPI) settingsHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/dynamic-ip-providers", handler: api.handleDynamicIPProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/in-user-rules", handler: api.handleInUserRules},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings/mihomo-native", handler: api.handleMihomoNativeConfig},
		{methods: []string{http.MethodGet}, path: "/settings/ip-fraud-providers", handler: api.handleIPFraudProviders},
		{methods: []string{http.MethodGet}, path: "/settings/ip-geo-providers", handler: api.handleIPGeoProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut}, path: "/settings", handler: api.handleRuntimeSettings},
	}
}
