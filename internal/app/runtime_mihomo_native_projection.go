package app

import (
	"context"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
)

func projectMihomoNativeSettings(ctx context.Context, settings mihomoNativeSettingsRepository, configDir string) error {
	if settings == nil {
		return nil
	}
	view, err := settings.loadMihomoNative(ctx)
	if err != nil {
		return err
	}
	if mihomonative.SettingsEmpty(view) {
		migrated, err := importMihomoNativeProjection(ctx, settings, configDir)
		if err != nil {
			return err
		}
		if !mihomonative.SettingsEmpty(migrated) {
			view = migrated
		}
	}
	if mihomonative.SettingsEmpty(view) && strings.TrimSpace(configDir) == "" {
		return nil
	}
	config, err := mihomonative.ConfigFromSettings(view)
	if err != nil {
		return err
	}
	return mihomonative.SaveConfig(configDir, config)
}
