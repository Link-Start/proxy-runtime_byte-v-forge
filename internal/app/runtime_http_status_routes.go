package app

import "net/http"

func (api *runtimeHTTPAPI) runtimeStatusHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet}, path: "/runtime/status", handler: api.handleRuntimeStatus},
	}
}
