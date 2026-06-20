package mihomonative

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func NormalizeSettings(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) *proxyruntimev1.ProxyRuntimeMihomoNativeConfig {
	if view == nil {
		view = &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{}
	}
	return &proxyruntimev1.ProxyRuntimeMihomoNativeConfig{
		FixedProxies:  normalizeFixedProxySettings(view.GetFixedProxies()),
		Subscriptions: normalizeSubscriptionSettings(view.GetSubscriptions()),
	}
}

func NormalizeFixedProxy(item FixedProxy, currentByName map[string]FixedProxy) FixedProxy {
	item.Name = strings.TrimSpace(item.Name)
	item.URI = strings.TrimSpace(item.URI)
	item.Type = strings.TrimSpace(item.Type)
	item.ID = appcore.RuntimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = appcore.RuntimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = StableID("fixed", item.URI)
	}
	return item
}

func NormalizeSubscription(item Subscription, currentByName map[string]Subscription) Subscription {
	item.Name = strings.TrimSpace(item.Name)
	item.URL = strings.TrimSpace(item.URL)
	item.ID = appcore.RuntimeSafeID(item.ID)
	if item.ID == "" && currentByName != nil {
		item.ID = appcore.RuntimeSafeID(currentByName[item.Name].ID)
	}
	if item.ID == "" {
		item.ID = StableID("sub", item.URL)
	}
	return item
}
