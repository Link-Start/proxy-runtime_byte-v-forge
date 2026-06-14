package mihomo

import "strings"

func mihomoProxyNames(proxies []map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, proxy := range proxies {
		name, _ := proxy["name"].(string)
		if name = strings.TrimSpace(name); name != "" {
			out[name] = struct{}{}
		}
	}
	return out
}

func mihomoProviderNames(providers map[string]mihomoProvider) map[string]struct{} {
	out := map[string]struct{}{}
	for name := range providers {
		if name = strings.TrimSpace(name); name != "" {
			out[name] = struct{}{}
		}
	}
	return out
}
