package app

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
