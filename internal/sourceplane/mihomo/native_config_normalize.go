package mihomo

import (
	"os"
	"path/filepath"
	"strings"
)

func normalizeNativeConfigPaths(configDir string, config *mihomoNativeConfig) {
	if config == nil || configDir == "" {
		return
	}
	for name, provider := range config.ProxyProviders {
		provider.Path = normalizeNativeProviderPath(configDir, provider.Path)
		normalizeNativeProviderHeaders(&provider)
		config.ProxyProviders[name] = provider
	}
	normalizeNativeGroups(config)
}

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

func normalizeNativeProviderPath(configDir string, providerPath string) string {
	providerPath = strings.TrimSpace(providerPath)
	if providerPath == "" {
		return ""
	}
	if !filepath.IsAbs(providerPath) {
		return filepath.Join(configDir, providerPath)
	}
	if rel, err := filepath.Rel(configDir, providerPath); err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".." {
		return providerPath
	}
	filename := strings.TrimSpace(filepath.Base(providerPath))
	if filename == "" || filename == "." || filename == string(os.PathSeparator) {
		return providerPath
	}
	return filepath.Join(configDir, "providers", filename)
}

func normalizeNativeProviderHeaders(provider *mihomoProvider) {
	if provider == nil {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(provider.Type), "http") || strings.TrimSpace(provider.URL) == "" {
		return
	}
	if strings.TrimSpace(provider.Proxy) == "" {
		provider.Proxy = defaultProviderFetchProxy
	}
	normalizeNativeProviderUserAgent(provider)
}

func normalizeNativeProviderUserAgent(provider *mihomoProvider) {
	if provider.Header == nil {
		provider.Header = map[string][]string{}
	}
	for key, values := range provider.Header {
		if !strings.EqualFold(strings.TrimSpace(key), "User-Agent") {
			continue
		}
		if replaceNativeProviderUserAgent(values) {
			provider.Header[key] = []string{defaultProviderUserAgent}
		}
		return
	}
	provider.Header["User-Agent"] = []string{defaultProviderUserAgent}
}

func replaceNativeProviderUserAgent(values []string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || strings.EqualFold(trimmed, legacyDefaultProviderUserAgent) {
			continue
		}
		return false
	}
	return true
}
