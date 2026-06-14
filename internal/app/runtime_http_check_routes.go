package app

import "net/http"

func (api *runtimeHTTPAPI) checkHTTPRoutes() []runtimeHTTPRoute {
	return []runtimeHTTPRoute{
		{methods: []string{http.MethodPost}, path: "/proxy_exit_ip", handler: api.handleGetProxyExitIP},
		{methods: []string{http.MethodPost}, path: "/proxy_exit_geo", handler: api.handleGetProxyExitGeo},
		{methods: []string{http.MethodPost}, path: "/ip_fraud_check", handler: api.handleCheckIPFraud},
		{methods: []string{http.MethodPost}, path: "/check_cf_access_risk", handler: api.handleCheckEdgeAccessRisk},
		{methods: []string{http.MethodPost}, path: "/target_connectivity_check", handler: api.handleCheckTargetConnectivity},
		{methods: []string{http.MethodGet, http.MethodPost}, path: "/proxy_exit_check_snapshot", handler: api.handleGetProxyExitCheckSnapshot},
	}
}
