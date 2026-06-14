package app

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

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
	if api.sessionAuthenticated(ctx.Request) {
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
	if pathInPrefix(requestPath, "/api/auth") {
		return false
	}
	return true
}

func authTokenMatches(actual string, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func (api *runtimeHTTPAPI) forwardMihomoControllerAuthorization(out *http.Request, in *http.Request) {
	if out == nil || in == nil {
		return
	}
	if in.URL == nil {
		return
	}
	if !pathInPrefix(in.URL.Path, "/mihomo/controller") {
		return
	}
	token := strings.TrimSpace(api.authToken)
	if token == "" {
		return
	}
	out.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	out.URL.RawQuery = mihomoControllerUpstreamRawQuery(out.URL.RawQuery)
}

func pathInPrefix(requestPath string, prefix string) bool {
	requestPath = strings.TrimRight(strings.TrimSpace(requestPath), "/")
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
}
