package auth

import (
	"net/http"
	"strings"
)

func Required(authToken string, requestPath string) bool {
	return RequestRequired(authToken, "", requestPath)
}

func RequestRequired(authToken string, method string, requestPath string) bool {
	if strings.TrimSpace(authToken) == "" {
		return false
	}
	requestPath = strings.TrimSpace(requestPath)
	switch strings.TrimRight(requestPath, "/") {
	case "", "/", "/healthz", "/readyz", "/metrics", "/login":
		return false
	}
	if publicServiceRuntimePath(method, requestPath) {
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

func publicServiceRuntimePath(method string, requestPath string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	path := strings.TrimRight(strings.TrimSpace(requestPath), "/")
	switch path {
	case "/api/leases/acquire", "/api/leases/release":
		return method == http.MethodPost
	case "/api/proxy_exit_ip",
		"/api/proxy_exit_geo",
		"/api/ip_fraud_check",
		"/api/check_cf_access_risk",
		"/api/target_connectivity_check":
		return method == http.MethodPost
	case "/api/proxy_exit_check_snapshot":
		return method == http.MethodGet || method == http.MethodPost
	default:
		return false
	}
}
