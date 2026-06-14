package app

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const bearerPrefix = "bearer "

func (api *runtimeHTTPAPI) authorize(ctx *gin.Context) bool {
	if !api.authRequired(ctx.Request.URL.Path) {
		return true
	}
	expected := strings.TrimSpace(api.authToken)
	if expected == "" {
		return true
	}
	if authenticatedRequest(ctx.Request, expected) {
		return true
	}
	ctx.Header("WWW-Authenticate", "Bearer")
	writeHTTPError(ctx.Writer, errors.New("unauthorized"), http.StatusUnauthorized)
	return false
}

func (api *runtimeHTTPAPI) authRequired(requestPath string) bool {
	if strings.TrimSpace(api.authToken) == "" {
		return false
	}
	requestPath = strings.TrimSpace(requestPath)
	return pathInPrefix(requestPath, controlPlaneHTTPPrefix) || pathInPrefix(requestPath, "/mihomo/controller")
}

func bearerToken(req *http.Request) string {
	if req == nil {
		return ""
	}
	value := strings.TrimSpace(req.Header.Get("Authorization"))
	if len(value) <= len(bearerPrefix) || !strings.EqualFold(value[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(value[len(bearerPrefix):])
}

func authenticatedRequest(req *http.Request, expected string) bool {
	if authTokenMatches(bearerToken(req), expected) {
		return true
	}
	if req == nil || req.URL == nil || !pathInPrefix(req.URL.Path, "/mihomo/controller") {
		return false
	}
	return authTokenMatches(mihomoControllerQueryToken(req), expected)
}

func authTokenMatches(actual string, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func mihomoControllerQueryToken(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return strings.TrimSpace(req.URL.Query().Get("token"))
}

func (api *runtimeHTTPAPI) forwardMihomoControllerAuthorization(out *http.Request, in *http.Request) {
	if out == nil || in == nil || out.Header.Get("Authorization") != "" {
		return
	}
	if in.URL == nil {
		return
	}
	if !pathInPrefix(in.URL.Path, "/mihomo/controller") {
		return
	}
	token := mihomoControllerQueryToken(in)
	if !authTokenMatches(token, api.authToken) {
		return
	}
	out.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
}

func pathInPrefix(requestPath string, prefix string) bool {
	requestPath = strings.TrimRight(strings.TrimSpace(requestPath), "/")
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
}
