package app

import "net/http"

func (api *runtimeHTTPAPI) leaseHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet}, path: "/leases", handler: api.handleLeases},
		{methods: []string{http.MethodPost}, path: "/leases/acquire", handler: api.handleAcquireLease},
		{methods: []string{http.MethodPost}, path: "/leases/release", handler: api.handleReleaseLease},
	}
}
