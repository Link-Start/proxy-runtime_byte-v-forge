package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func mihomoNativeConfigPath(runtime *Runtime) (string, error) {
	if runtime == nil {
		return "", errors.New("runtime is required")
	}
	return mihomoNativeConfigPathFromDir(runtime.cfg.Mihomo.ConfigDir)
}

func mihomoNativeConfigPathFromDir(configDir string) (string, error) {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return "", errors.New("mihomo config dir is required")
	}
	return filepath.Join(configDir, mihomoNativeFileName), nil
}

func loadMihomoNativeProjection(runtime *Runtime) (mihomoNativeConfigFile, bool, error) {
	path, err := mihomoNativeConfigPath(runtime)
	if err != nil {
		return mihomoNativeConfigFile{}, false, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return mihomoNativeConfigFile{}, false, nil
	}
	if err != nil {
		return mihomoNativeConfigFile{}, false, fmt.Errorf("read mihomo native config projection: %w", err)
	}
	if len(data) == 0 {
		return mihomoNativeConfigFile{}, false, nil
	}
	var config mihomoNativeConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return mihomoNativeConfigFile{}, false, fmt.Errorf("parse mihomo native config projection: %w", err)
	}
	if config.ProxyProviders == nil {
		config.ProxyProviders = map[string]mihomoNativeProvider{}
	}
	return config, true, nil
}

func saveMihomoNativeConfig(runtime *Runtime, config mihomoNativeConfigFile) error {
	path, err := mihomoNativeConfigPath(runtime)
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
