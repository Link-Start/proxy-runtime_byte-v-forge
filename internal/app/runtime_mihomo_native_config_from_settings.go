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
	for _, item := range view.GetFixedProxies() {
		proxy := nativeFixedProxyFromProto(item)
		rendered, err := mihomoNativeProxyFromURI(proxy.Name, proxy.URI)
		if err != nil {
			return mihomoNativeConfigFile{}, err
		}
		proxy.Type = jsonStringValue(rendered["type"])
		config.FixedProxies = append(config.FixedProxies, proxy)
		config.Proxies = append(config.Proxies, rendered)
	}
	for _, item := range view.GetSubscriptions() {
		provider, subscription, err := mihomoNativeSubscriptionProvider(nativeSubscriptionFromProto(item))
		if err != nil {
			return mihomoNativeConfigFile{}, err
		}
		config.Subscriptions = append(config.Subscriptions, subscription)
		config.ProxyProviders[subscription.Name] = provider
	}
	return config, nil
}
