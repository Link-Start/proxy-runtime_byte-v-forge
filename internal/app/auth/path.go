package auth

import "strings"

func Required(authToken string, requestPath string) bool {
	if strings.TrimSpace(authToken) == "" {
		return false
	}
	requestPath = strings.TrimSpace(requestPath)
	switch strings.TrimRight(requestPath, "/") {
	case "", "/", "/healthz", "/readyz", "/metrics", "/login":
		return false
	}
	return !PublicRuntimePath(requestPath)
}

func PublicRuntimePath(requestPath string) bool {
	switch strings.TrimRight(strings.TrimSpace(requestPath), "/") {
	case "/api/auth/session", "/api/auth/login", "/api/auth/logout":
		return true
	default:
		return false
	}
}
