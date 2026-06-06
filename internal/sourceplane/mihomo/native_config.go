package mihomo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const nativeConfigFileName = "native.json"

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

func normalizeNativeConfigPaths(configDir string, config *mihomoNativeConfig) {
	if config == nil || configDir == "" {
		return
	}
	for name, provider := range config.ProxyProviders {
		if provider.Path != "" && !filepath.IsAbs(provider.Path) {
			provider.Path = filepath.Join(configDir, provider.Path)
			config.ProxyProviders[name] = provider
		}
	}
}

func cloneNativeProxies(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		copy := make(map[string]any, len(item))
		for key, value := range item {
			copy[key] = value
		}
		out = append(out, copy)
	}
	return out
}

func cloneNativeProviders(items map[string]mihomoProvider) map[string]mihomoProvider {
	out := make(map[string]mihomoProvider, len(items))
	for key, item := range items {
		if key == "" {
			continue
		}
		out[key] = item
	}
	return out
}

func cloneNativeGroups(items []mihomoGroup) []mihomoGroup {
	out := make([]mihomoGroup, 0, len(items))
	for _, item := range items {
		if item.Name == "" {
			continue
		}
		item.Proxies = append([]string(nil), item.Proxies...)
		item.Use = append([]string(nil), item.Use...)
		out = append(out, item)
	}
	return out
}
