package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

func (api *runtimeHTTPAPI) providerHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/providers", Handler: api.handleProviders},
		{Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, Path: "/provider-accounts", Handler: api.handleProviderAccounts},
	}
}
