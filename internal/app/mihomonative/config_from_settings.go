package mihomonative

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func ConfigFromSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (ConfigFile, error) {
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
