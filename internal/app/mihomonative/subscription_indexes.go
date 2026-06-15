package mihomonative

import "strings"

func CurrentSubscriptionIndexes(config ConfigFile) (map[string]Subscription, map[string]Subscription) {
	byID := map[string]Subscription{}
	byName := map[string]Subscription{}
	for _, item := range config.Subscriptions {
		normalized := NormalizeSubscription(item, nil)
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
		normalized := Subscription{ID: StableID("sub", url), Name: strings.TrimSpace(name), URL: url}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}
