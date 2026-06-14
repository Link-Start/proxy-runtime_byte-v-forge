package mihomo

func cloneNativeProxies(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		copy := make(map[string]any, len(item))
		for key, value := range item {
			copy[key] = value
		}
		out = append(out, copy)
	}
	return out
}

func cloneNativeProviders(items map[string]mihomoProvider) map[string]mihomoProvider {
	out := make(map[string]mihomoProvider, len(items))
	for key, item := range items {
		if key == "" {
			continue
		}
		out[key] = item
	}
	return out
}

func cloneNativeGroups(items []mihomoGroup) []mihomoGroup {
	out := make([]mihomoGroup, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		item.Proxies = append([]string(nil), item.Proxies...)
		item.Use = append([]string(nil), item.Use...)
		out = append(out, item)
	}
	return out
}
