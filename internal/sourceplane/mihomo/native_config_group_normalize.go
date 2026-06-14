package mihomo

import "strings"

func normalizeNativeGroups(config *mihomoNativeConfig) {
	if config == nil || len(config.ProxyGroups) == 0 {
		return
	}
	groups := make([]mihomoGroup, 0, len(config.ProxyGroups))
	for _, group := range config.ProxyGroups {
		if isNativeFixedProxyGroup(group) {
			continue
		}
		groups = append(groups, group)
	}
	config.ProxyGroups = groups
}

func isNativeFixedProxyGroup(group mihomoGroup) bool {
	if strings.TrimSpace(group.Name) != nativeFixedProxyGroupName {
		return false
	}
	if strings.TrimSpace(group.Type) != "" && !strings.EqualFold(strings.TrimSpace(group.Type), nativeFixedProxyGroupKind) {
		return false
	}
	return len(group.Use) == 0 && len(group.Proxies) > 0 && group.Proxies[0] != nativeFixedProxyGroupProxy
}
