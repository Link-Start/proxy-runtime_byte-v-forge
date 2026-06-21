package mihomonative

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func ConfigFromSettings(view *proxygatewayv1.ProxyGatewayMihomoNativeConfig) (ConfigFile, error) {
	view = NormalizeSettings(view)
	config := ConfigFile{
		FixedProxies:   make([]FixedProxy, 0, len(view.GetFixedProxies())),
		Proxies:        make([]map[string]any, 0, len(view.GetFixedProxies())),
		Subscriptions:  make([]Subscription, 0, len(view.GetSubscriptions())),
		ProxyProviders: map[string]Provider{},
	}
	if err := appendMihomoNativeFixedProxyConfig(&config, view.GetFixedProxies()); err != nil {
		return ConfigFile{}, err
	}
	if err := appendMihomoNativeSubscriptionConfig(&config, view.GetSubscriptions()); err != nil {
		return ConfigFile{}, err
	}
	return config, nil
}
