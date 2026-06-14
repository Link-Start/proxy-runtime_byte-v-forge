package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

func (api *runtimeHTTPAPI) runtimeStatusHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/runtime/status", Handler: api.handleRuntimeStatus},
	}
}
