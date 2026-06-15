package mihomonative

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func FixedProxySettingsFromConfig(config ConfigFile) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	if len(config.FixedProxies) > 0 {
		return FixedProxySettingsFromExplicitConfig(config.FixedProxies)
	}
	return FixedProxySettingsFromNativeProxies(config.Proxies)
}

func FixedProxySettingsFromExplicitConfig(proxies []FixedProxy) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(proxies))
	for _, proxy := range proxies {
		item := NormalizeFixedProxy(proxy, nil)
		if item.Name == "" || item.URI == "" {
			continue
		}
		out = append(out, protoFixedProxy(item))
	}
	return out
}

func FixedProxySettingsFromNativeProxies(proxies []map[string]any) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(proxies))
	for _, proxy := range proxies {
		name := jsonStringValue(proxy["name"])
		proxyType := jsonStringValue(proxy["type"])
		if name == "" {
			continue
		}
		uri := proxyURI(proxy)
		if uri == "" {
			continue
		}
		out = append(out, protoFixedProxy(FixedProxy{ID: StableID("fixed", uri), Name: name, Type: proxyType, URI: uri}))
	}
	return out
}

func SubscriptionSettingsFromConfig(config ConfigFile) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	if len(config.Subscriptions) > 0 {
		return SubscriptionSettingsFromExplicitConfig(config.Subscriptions)
	}
	return SubscriptionSettingsFromProviders(config.ProxyProviders)
}

func SubscriptionSettingsFromExplicitConfig(subscriptions []Subscription) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		item := NormalizeSubscription(subscription, nil)
		if item.Name == "" || item.URL == "" {
			continue
		}
		out = append(out, protoSubscription(item))
	}
	return out
}

func SubscriptionSettingsFromProviders(providers map[string]Provider) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(providers))
	for name, provider := range providers {
		if strings.TrimSpace(provider.URL) == "" {
			continue
		}
		out = append(out, protoSubscription(Subscription{ID: StableID("sub", provider.URL), Name: name, URL: provider.URL}))
	}
	return out
}
