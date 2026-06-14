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
