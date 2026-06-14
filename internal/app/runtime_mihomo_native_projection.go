package app

import (
	"context"
	"strings"
)

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
