package mihomonative

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConfigPathFromDir(configDir string) (string, error) {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return "", errors.New("mihomo config dir is required")
	}
	return filepath.Join(configDir, fileName), nil
}

func LoadProjection(configDir string) (ConfigFile, bool, error) {
	path, err := ConfigPathFromDir(configDir)
	if err != nil {
		return ConfigFile{}, false, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ConfigFile{}, false, nil
	}
	if err != nil {
		return ConfigFile{}, false, fmt.Errorf("read mihomo native config projection: %w", err)
	}
	if len(data) == 0 {
		return ConfigFile{}, false, nil
	}
	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return ConfigFile{}, false, fmt.Errorf("parse mihomo native config projection: %w", err)
	}
	if config.ProxyProviders == nil {
		config.ProxyProviders = map[string]Provider{}
	}
	return config, true, nil
}

func SaveConfig(configDir string, config ConfigFile) error {
	path, err := ConfigPathFromDir(configDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create mihomo config dir: %w", err)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mihomo native config: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".native-*.json")
	if err != nil {
		return fmt.Errorf("create mihomo native config file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write mihomo native config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close mihomo native config: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return fmt.Errorf("chmod mihomo native config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace mihomo native config: %w", err)
	}
	return nil
}
