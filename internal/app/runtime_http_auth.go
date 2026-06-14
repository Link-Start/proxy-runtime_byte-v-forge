package app

import (
	"errors"
	"net/http"
	"strings"

	authapp "github.com/byte-v-forge/proxy-runtime/internal/app/auth"
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
