package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

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
