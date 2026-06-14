package app

import "net/http"

func (api *runtimeHTTPAPI) authHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodGet}, path: "/auth/ws-token", handler: api.handleAuthWebSocketToken},
	}
}
