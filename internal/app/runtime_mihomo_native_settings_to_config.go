package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func appendMihomoNativeFixedProxyConfig(config *mihomoNativeConfigFile, items []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) error {
	for _, item := range items {
		proxy := nativeFixedProxyFromProto(item)
		rendered, err := mihomoNativeProxyFromURI(proxy.Name, proxy.URI)
		if err != nil {
			return err
		}
		proxy.Type = jsonStringValue(rendered["type"])
		config.FixedProxies = append(config.FixedProxies, proxy)
		config.Proxies = append(config.Proxies, rendered)
	}
	return nil
}

func appendMihomoNativeSubscriptionConfig(config *mihomoNativeConfigFile, items []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) error {
	for _, item := range items {
		provider, subscription, err := mihomoNativeSubscriptionProvider(nativeSubscriptionFromProto(item))
		if err != nil {
			return err
		}
		config.Subscriptions = append(config.Subscriptions, subscription)
		config.ProxyProviders[subscription.Name] = provider
	}
	return nil
}
