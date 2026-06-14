package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type mihomoNativeSettingsRepository interface {
	loadMihomoNative(context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error)
	saveMihomoNative(context.Context, *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) error
}

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	return projectMihomoNativeSettings(ctx, r.settings, r.cfg.Mihomo.ConfigDir)
}

func projectMihomoNativeSettings(ctx context.Context, settings mihomoNativeSettingsRepository, configDir string) error {
	if settings == nil {
		return nil
	}
	view, err := settings.loadMihomoNative(ctx)
	if err != nil {
		return err
	}
	if mihomoNativeSettingsEmpty(view) {
		migrated, err := importMihomoNativeProjection(ctx, settings, configDir)
		if err != nil {
			return err
		}
		if !mihomoNativeSettingsEmpty(migrated) {
			view = migrated
		}
	}
	if mihomoNativeSettingsEmpty(view) && strings.TrimSpace(configDir) == "" {
		return nil
	}
	config, err := mihomoNativeConfigFileFromSettings(view)
	if err != nil {
		return err
	}
	return saveMihomoNativeConfig(configDir, config)
}

func importMihomoNativeProjection(ctx context.Context, settings mihomoNativeSettingsRepository, configDir string) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	config, exists, err := loadMihomoNativeProjection(configDir)
	if err != nil || !exists {
		return nil, err
	}
	view := mihomoNativeSettingsFromConfig(config)
	if mihomoNativeSettingsEmpty(view) {
		return nil, nil
	}
	if err := settings.saveMihomoNative(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}
