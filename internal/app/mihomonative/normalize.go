package mihomonative

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func NormalizeSettings(view *proxygatewayv1.ProxyGatewayMihomoNativeConfig) *proxygatewayv1.ProxyGatewayMihomoNativeConfig {
	if view == nil {
		view = &proxygatewayv1.ProxyGatewayMihomoNativeConfig{}
	}
	return &proxygatewayv1.ProxyGatewayMihomoNativeConfig{
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
