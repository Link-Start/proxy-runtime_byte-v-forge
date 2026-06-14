package mihomo

import (
	"os"
	"path/filepath"
	"strings"
)

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
