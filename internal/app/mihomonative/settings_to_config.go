package mihomonative

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func appendMihomoNativeFixedProxyConfig(config *ConfigFile, items []*proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy) error {
	for _, item := range items {
		proxy := FixedProxyFromProto(item)
		rendered, err := proxyFromURI(proxy.Name, proxy.URI)
		if err != nil {
			return err
		}
		proxy.Type = jsonStringValue(rendered["type"])
		config.FixedProxies = append(config.FixedProxies, proxy)
		config.Proxies = append(config.Proxies, rendered)
	}
	return nil
}

func appendMihomoNativeSubscriptionConfig(config *ConfigFile, items []*proxygatewayv1.ProxyGatewayMihomoNativeSubscription) error {
	for _, item := range items {
		provider, subscription, err := SubscriptionProvider(SubscriptionFromProto(item))
		if err != nil {
			return err
		}
		config.Subscriptions = append(config.Subscriptions, subscription)
		config.ProxyProviders[subscription.Name] = provider
	}
	return nil
}
