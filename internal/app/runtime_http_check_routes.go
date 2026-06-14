package app

import (
	httpapi "github.com/byte-v-forge/proxy-runtime/internal/app/httpapi"
	"net/http"
)

func (api *runtimeHTTPAPI) checkHTTPRoutes() []httpapi.Route {
	return []httpapi.Route{
		{Methods: []string{http.MethodPost}, Path: "/proxy_exit_ip", Handler: api.handleGetProxyExitIP},
		{Methods: []string{http.MethodPost}, Path: "/proxy_exit_geo", Handler: api.handleGetProxyExitGeo},
		{Methods: []string{http.MethodPost}, Path: "/ip_fraud_check", Handler: api.handleCheckIPFraud},
		{Methods: []string{http.MethodPost}, Path: "/check_cf_access_risk", Handler: api.handleCheckEdgeAccessRisk},
		{Methods: []string{http.MethodPost}, Path: "/target_connectivity_check", Handler: api.handleCheckTargetConnectivity},
		{Methods: []string{http.MethodGet, http.MethodPost}, Path: "/proxy_exit_check_snapshot", Handler: api.handleGetProxyExitCheckSnapshot},
	}
}
