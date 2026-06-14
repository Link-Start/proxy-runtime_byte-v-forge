package app

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *runtimeHTTPAPI) authorize(ctx *gin.Context) bool {
	if !api.authRequired(ctx.Request.URL.Path) {
		return true
	}
	if !api.auth.Enabled() {
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
	return api.auth.Required(requestPath)
}
