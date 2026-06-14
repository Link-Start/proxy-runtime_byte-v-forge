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
	actual := bearerToken(ctx.Request)
	if subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1 {
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

func pathInPrefix(requestPath string, prefix string) bool {
	requestPath = strings.TrimRight(strings.TrimSpace(requestPath), "/")
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
}
