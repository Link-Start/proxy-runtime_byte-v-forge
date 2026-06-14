package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	view, err := r.settings.loadMihomoNative(ctx)
	if err != nil {
		return err
	}
	if mihomoNativeSettingsEmpty(view) {
		migrated, err := r.importMihomoNativeProjection(ctx)
		if err != nil {
			return err
		}
		if !mihomoNativeSettingsEmpty(migrated) {
			view = migrated
		}
	}
	if mihomoNativeSettingsEmpty(view) && strings.TrimSpace(r.cfg.Mihomo.ConfigDir) == "" {
		return nil
	}
	config, err := mihomoNativeConfigFileFromSettings(view)
	if err != nil {
		return err
	}
	return saveMihomoNativeConfig(r.cfg.Mihomo.ConfigDir, config)
}

func (r *Runtime) importMihomoNativeProjection(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	config, exists, err := loadMihomoNativeProjection(r.cfg.Mihomo.ConfigDir)
	if err != nil || !exists {
		return nil, err
	}
	view := mihomoNativeSettingsFromConfig(config)
	if mihomoNativeSettingsEmpty(view) {
		return nil, nil
	}
	if err := r.settings.saveMihomoNative(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}
