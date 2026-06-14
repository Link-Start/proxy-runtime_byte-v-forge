package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func mihomoNativeFixedProxySettingsFromConfig(config mihomoNativeConfigFile) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	if len(config.FixedProxies) > 0 {
		return mihomoNativeFixedProxySettingsFromExplicitConfig(config.FixedProxies)
	}
	return mihomoNativeFixedProxySettingsFromNativeProxies(config.Proxies)
}

func mihomoNativeFixedProxySettingsFromExplicitConfig(proxies []mihomoNativeFixedProxy) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(proxies))
	for _, proxy := range proxies {
		item := normalizeMihomoNativeFixedProxy(proxy, nil)
		if item.Name == "" || item.URI == "" {
			continue
		}
		out = append(out, protoMihomoNativeFixedProxy(item))
	}
	return out
}

func mihomoNativeFixedProxySettingsFromNativeProxies(proxies []map[string]any) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(proxies))
	for _, proxy := range proxies {
		name := jsonStringValue(proxy["name"])
		proxyType := jsonStringValue(proxy["type"])
		if name == "" {
			continue
		}
		uri := mihomoNativeProxyURI(proxy)
		if uri == "" {
			continue
		}
		out = append(out, protoMihomoNativeFixedProxy(mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: proxyType, URI: uri}))
	}
	return out
}

func mihomoNativeSubscriptionSettingsFromConfig(config mihomoNativeConfigFile) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	if len(config.Subscriptions) > 0 {
		return mihomoNativeSubscriptionSettingsFromExplicitConfig(config.Subscriptions)
	}
	return mihomoNativeSubscriptionSettingsFromProviders(config.ProxyProviders)
}

func mihomoNativeSubscriptionSettingsFromExplicitConfig(subscriptions []mihomoNativeSubscription) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		item := normalizeMihomoNativeSubscription(subscription, nil)
		if item.Name == "" || item.URL == "" {
			continue
		}
		out = append(out, protoMihomoNativeSubscription(item))
	}
	return out
}

func mihomoNativeSubscriptionSettingsFromProviders(providers map[string]mihomoNativeProvider) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(providers))
	for name, provider := range providers {
		if strings.TrimSpace(provider.URL) == "" {
			continue
		}
		out = append(out, protoMihomoNativeSubscription(mihomoNativeSubscription{ID: nativeStableID("sub", provider.URL), Name: name, URL: provider.URL}))
	}
	return out
}
