package app

import (
	"errors"
	"net/http"
	"strings"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
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
	return authapp.Required(api.authToken, requestPath)
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
