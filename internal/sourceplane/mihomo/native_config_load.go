package mihomo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func loadNativeConfig(configDir string) (mihomoNativeConfig, error) {
	path := filepath.Join(configDir, nativeConfigFileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mihomoNativeConfig{}, nil
	}
	if err != nil {
		return mihomoNativeConfig{}, fmt.Errorf("read mihomo native config: %w", err)
	}
	if len(data) == 0 {
		return mihomoNativeConfig{}, nil
	}
	var config mihomoNativeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return mihomoNativeConfig{}, fmt.Errorf("parse mihomo native config: %w", err)
	}
	normalizeNativeConfigPaths(configDir, &config)
	return config, nil
}
