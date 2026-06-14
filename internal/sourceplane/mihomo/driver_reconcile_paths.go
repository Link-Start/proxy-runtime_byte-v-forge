package mihomo

import (
	"os"
	"path/filepath"
)

func ensureProviderConfigDir(configDir string) error {
	return os.MkdirAll(filepath.Join(configDir, "providers"), 0o700)
}

func runtimeConfigPath(configDir string) string {
	return filepath.Join(configDir, "config.json")
}
