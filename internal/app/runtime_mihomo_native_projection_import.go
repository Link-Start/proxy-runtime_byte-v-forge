package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
