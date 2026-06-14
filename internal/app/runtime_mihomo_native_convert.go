package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func mihomoNativeSettingsFromConfig(config mihomoNativeConfigFile) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	view := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	if len(config.FixedProxies) > 0 {
		for _, proxy := range config.FixedProxies {
			item := normalizeMihomoNativeFixedProxy(proxy, nil)
			if item.Name == "" || item.URI == "" {
				continue
			}
			view.FixedProxies = append(view.FixedProxies, protoMihomoNativeFixedProxy(item))
		}
	} else {
		for _, proxy := range config.Proxies {
			name := jsonStringValue(proxy["name"])
			proxyType := jsonStringValue(proxy["type"])
			if name == "" {
				continue
			}
			uri := mihomoNativeProxyURI(proxy)
			if uri == "" {
				continue
			}
			view.FixedProxies = append(view.FixedProxies, protoMihomoNativeFixedProxy(mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: proxyType, URI: uri}))
		}
	}
	if len(config.Subscriptions) > 0 {
		for _, subscription := range config.Subscriptions {
			item := normalizeMihomoNativeSubscription(subscription, nil)
			if item.Name == "" || item.URL == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, protoMihomoNativeSubscription(item))
		}
	} else {
		for name, provider := range config.ProxyProviders {
			if strings.TrimSpace(provider.URL) == "" {
				continue
			}
			view.Subscriptions = append(view.Subscriptions, protoMihomoNativeSubscription(mihomoNativeSubscription{ID: nativeStableID("sub", provider.URL), Name: name, URL: provider.URL}))
		}
	}
	return normalizeMihomoNativeSettings(view)
}
