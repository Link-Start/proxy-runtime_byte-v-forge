package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

func (api *runtimeHTTPAPI) leaseHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/leases", Handler: api.handleLeases},
		{Methods: []string{http.MethodPost}, Path: "/leases/acquire", Handler: api.handleAcquireLease},
		{Methods: []string{http.MethodPost}, Path: "/leases/release", Handler: api.handleReleaseLease},
	}
}
