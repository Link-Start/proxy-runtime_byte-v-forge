package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func mihomoNativeConfigFileFromSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (mihomoNativeConfigFile, error) {
	view = normalizeMihomoNativeSettings(view)
	config := mihomoNativeConfigFile{
		FixedProxies:   make([]mihomoNativeFixedProxy, 0, len(view.GetFixedProxies())),
		Proxies:        make([]map[string]any, 0, len(view.GetFixedProxies())),
		Subscriptions:  make([]mihomoNativeSubscription, 0, len(view.GetSubscriptions())),
		ProxyProviders: map[string]mihomoNativeProvider{},
	}
	if err := appendMihomoNativeFixedProxyConfig(&config, view.GetFixedProxies()); err != nil {
		return mihomoNativeConfigFile{}, err
	}
	if err := appendMihomoNativeSubscriptionConfig(&config, view.GetSubscriptions()); err != nil {
		return mihomoNativeConfigFile{}, err
	}
	return config, nil
}
