package mihomonative

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func SettingsEmpty(view *proxygatewayv1.ProxyGatewayMihomoNativeConfig) bool {
	return view == nil || len(view.GetFixedProxies()) == 0 && len(view.GetSubscriptions()) == 0
}
