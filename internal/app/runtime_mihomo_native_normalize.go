package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func normalizeMihomoNativeSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	return &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  normalizeMihomoNativeFixedProxySettings(view.GetFixedProxies()),
		Subscriptions: normalizeMihomoNativeSubscriptionSettings(view.GetSubscriptions()),
	}
}

func normalizeMihomoNativeFixedProxy(item mihomoNativeFixedProxy, currentByName map[string]mihomoNativeFixedProxy) mihomoNativeFixedProxy {
	item.Name = strings.TrimSpace(item.Name)
	item.URI = strings.TrimSpace(item.URI)
	item.Type = strings.TrimSpace(item.Type)
	item.ID = runtimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = runtimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = nativeStableID("fixed", item.URI)
	}
	return item
}

func normalizeMihomoNativeSubscription(item mihomoNativeSubscription, currentByName map[string]mihomoNativeSubscription) mihomoNativeSubscription {
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	item.ID = runtimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = runtimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = nativeStableID("sub", item.URL)
	}
	return item
}
