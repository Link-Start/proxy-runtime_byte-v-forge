package mihomo

import "strings"

func mihomoProxyAvailable(available map[string]struct{}, name string) bool {
	if len(available) == 0 {
		return false
	}
	_, exists := available[strings.TrimSpace(name)]
	return exists
}

func mihomoProviderAvailable(available map[string]struct{}, name string) bool {
	if len(available) == 0 {
		return false
	}
	_, exists := available[strings.TrimSpace(name)]
	return exists
}
