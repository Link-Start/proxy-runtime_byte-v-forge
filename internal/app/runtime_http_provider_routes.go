package app

import "net/http"

func (api *runtimeHTTPAPI) providerHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet}, path: "/providers", handler: api.handleProviders},
		{methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, path: "/provider-accounts", handler: api.handleProviderAccounts},
	}
}
