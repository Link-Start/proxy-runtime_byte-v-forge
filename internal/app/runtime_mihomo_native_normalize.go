package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func normalizeMihomoNativeSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	out := &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(view.GetFixedProxies())),
		Subscriptions: make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(view.GetSubscriptions())),
	}
	seenFixed := map[string]struct{}{}
	for _, item := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seenFixed[key]; exists {
			continue
		}
		seenFixed[key] = struct{}{}
		out.FixedProxies = append(out.FixedProxies, protoMihomoNativeFixedProxy(normalized))
	}
	seenSubscriptions := map[string]struct{}{}
	for _, item := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seenSubscriptions[key]; exists {
			continue
		}
		seenSubscriptions[key] = struct{}{}
		out.Subscriptions = append(out.Subscriptions, protoMihomoNativeSubscription(normalized))
	}
	return out
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
