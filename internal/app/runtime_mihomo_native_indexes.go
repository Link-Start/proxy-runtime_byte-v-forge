package app

import (
	"strings"
)

func preserveNativeGroups(groups []mihomoNativeGroup) []mihomoNativeGroup {
	out := make([]mihomoNativeGroup, 0, len(groups))
	for _, group := range groups {
		if strings.TrimSpace(group.Name) == "" || group.Name == mihomoFixedProxyGroupName {
			continue
		}
		out = append(out, group)
	}
	return out
}
func currentFixedProxyIndexes(config mihomoNativeConfigFile) (map[string]mihomoNativeFixedProxy, map[string]mihomoNativeFixedProxy) {
	byID := map[string]mihomoNativeFixedProxy{}
	byName := map[string]mihomoNativeFixedProxy{}
	for _, item := range config.FixedProxies {
		normalized := normalizeMihomoNativeFixedProxy(item, nil)
		if normalized.ID != "" {
			byID[normalized.ID] = normalized
		}
		if normalized.Name != "" {
			byName[normalized.Name] = normalized
		}
	}
	for _, proxy := range config.Proxies {
		name := jsonStringValue(proxy["name"])
		uri := mihomoNativeProxyURI(proxy)
		if name == "" || uri == "" {
			continue
		}
		normalized := mihomoNativeFixedProxy{ID: nativeStableID("fixed", uri), Name: name, Type: jsonStringValue(proxy["type"]), URI: uri}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}
func currentSubscriptionIndexes(config mihomoNativeConfigFile) (map[string]mihomoNativeSubscription, map[string]mihomoNativeSubscription) {
	byID := map[string]mihomoNativeSubscription{}
	byName := map[string]mihomoNativeSubscription{}
	for _, item := range config.Subscriptions {
		normalized := normalizeMihomoNativeSubscription(item, nil)
		if normalized.ID != "" {
			byID[normalized.ID] = normalized
		}
		if normalized.Name != "" {
			byName[normalized.Name] = normalized
		}
	}
	for name, provider := range config.ProxyProviders {
		url := strings.TrimSpace(provider.URL)
		if strings.TrimSpace(name) == "" || url == "" {
			continue
		}
		normalized := mihomoNativeSubscription{ID: nativeStableID("sub", url), Name: strings.TrimSpace(name), URL: url}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}
