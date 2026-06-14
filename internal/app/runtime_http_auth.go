package app

import (
	"errors"
	"net/http"
	"strings"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) authorize(ctx *gin.Context) bool {
	if !api.authRequired(ctx.Request.URL.Path) {
		return true
	}
	expected := strings.TrimSpace(api.authToken)
	if expected == "" {
		return true
	}
	if api.requestAuthenticated(ctx.Request) {
		return true
	}
	if api.redirectLoginPreferred(ctx.Request) {
		api.redirectToLogin(ctx, ctx.Request.URL.RequestURI())
		return false
	}
	ctx.Header("WWW-Authenticate", "Cookie")
	writeHTTPError(ctx.Writer, errors.New("unauthorized"), http.StatusUnauthorized)
	return false
}

func (api *runtimeHTTPAPI) authRequired(requestPath string) bool {
	if strings.TrimSpace(api.authToken) == "" {
		return false
	}
	requestPath = strings.TrimSpace(requestPath)
	switch strings.TrimRight(requestPath, "/") {
	case "", "/", "/healthz", "/readyz", "/login":
		return false
	}
	if publicRuntimeAuthPath(requestPath) {
		return false
	}
	return true
}

func (api *runtimeHTTPAPI) forwardMihomoControllerAuthorization(out *http.Request, requestPath string) {
	if out == nil {
		return
	}
	if !httpapi.PathInPrefix(requestPath, "/mihomo/controller") {
		return
	}
	token := strings.TrimSpace(api.authToken)
	if token == "" {
		return
	}
	out.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	out.URL.RawQuery = dashboardapp.ControllerUpstreamRawQuery(out.URL.RawQuery)
}

func publicRuntimeAuthPath(requestPath string) bool {
	switch strings.TrimRight(strings.TrimSpace(requestPath), "/") {
	case "/api/auth/session", "/api/auth/login", "/api/auth/logout":
		return true
	default:
		return false
	}
}
