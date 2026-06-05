package mihomo

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeConfigData(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create mihomo config dir: %w", err)
	}
	file, err := os.CreateTemp(dir, ".mihomo-*.json")
	if err != nil {
		return fmt.Errorf("create mihomo config file: %w", err)
	}
	tempPath := file.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write mihomo config file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close mihomo config file: %w", err)
	}
	if err := os.Chmod(tempPath, 0o600); err != nil {
		return fmt.Errorf("chmod mihomo config file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace mihomo config file: %w", err)
	}
	return nil
}
