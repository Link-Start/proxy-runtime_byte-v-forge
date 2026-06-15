package auth

import (
	"net/http"
	"strings"
)

func (a Application) serviceRequestAuthenticated(req *http.Request) bool {
	if a.serviceSecret == "" || req == nil || req.URL == nil {
		return false
	}
	if !ServiceRuntimePath(req.Method, req.URL.Path) {
		return false
	}
	return TokenMatches(bearerToken(req.Header.Get("Authorization")), a.serviceSecret)
}

func bearerToken(value string) string {
	scheme, token, ok := strings.Cut(strings.TrimSpace(value), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
