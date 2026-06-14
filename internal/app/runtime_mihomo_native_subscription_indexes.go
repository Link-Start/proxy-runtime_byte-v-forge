package app

import "strings"

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
