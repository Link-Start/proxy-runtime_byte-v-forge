package mihomonative

func CurrentFixedProxyIndexes(config ConfigFile) (map[string]FixedProxy, map[string]FixedProxy) {
	byID := map[string]FixedProxy{}
	byName := map[string]FixedProxy{}
	for _, item := range config.FixedProxies {
		normalized := NormalizeFixedProxy(item, nil)
		if normalized.ID != "" {
			byID[normalized.ID] = normalized
		}
		if normalized.Name != "" {
			byName[normalized.Name] = normalized
		}
	}
	for _, proxy := range config.Proxies {
		name := jsonStringValue(proxy["name"])
		uri := proxyURI(proxy)
		if name == "" || uri == "" {
			continue
		}
		normalized := FixedProxy{ID: StableID("fixed", uri), Name: name, Type: jsonStringValue(proxy["type"]), URI: uri}
		if _, exists := byID[normalized.ID]; !exists {
			byID[normalized.ID] = normalized
		}
		if _, exists := byName[normalized.Name]; !exists {
			byName[normalized.Name] = normalized
		}
	}
	return byID, byName
}
