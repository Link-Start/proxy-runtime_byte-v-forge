package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

func (api *runtimeHTTPAPI) authHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodGet}, Path: "/auth/ws-token", Handler: api.handleAuthWebSocketToken},
	}
}
