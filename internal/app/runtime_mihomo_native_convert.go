package app

import (
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
func protoMihomoNativeFixedProxy(item mihomoNativeFixedProxy) *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy{Id: item.ID, Name: item.Name, Type: item.Type, Uri: item.URI}
}
func protoMihomoNativeSubscription(item mihomoNativeSubscription) *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeSubscription{Id: item.ID, Name: item.Name, Url: item.URL}
}
func nativeFixedProxyFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) mihomoNativeFixedProxy {
	if item == nil {
		return mihomoNativeFixedProxy{}
	}
	return mihomoNativeFixedProxy{ID: item.GetId(), Name: item.GetName(), Type: item.GetType(), URI: item.GetUri()}
}
func nativeSubscriptionFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) mihomoNativeSubscription {
	if item == nil {
		return mihomoNativeSubscription{}
	}
	return mihomoNativeSubscription{ID: item.GetId(), Name: item.GetName(), URL: item.GetUrl()}
}
