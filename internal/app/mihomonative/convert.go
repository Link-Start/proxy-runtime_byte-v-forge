package mihomonative

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func SettingsFromConfig(config ConfigFile) *proxygatewayv1.ProxyGatewayMihomoNativeConfig {
	view := &proxygatewayv1.ProxyGatewayMihomoNativeConfig{
		FixedProxies:  FixedProxySettingsFromConfig(config),
		Subscriptions: SubscriptionSettingsFromConfig(config),
	}
	return NormalizeSettings(view)
}
