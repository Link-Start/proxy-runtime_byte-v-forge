package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
)

func importMihomoNativeProjection(ctx context.Context, settings mihomoNativeSettingsRepository, configDir string) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	config, exists, err := mihomonative.LoadProjection(configDir)
	if err != nil || !exists {
		return nil, err
	}
	view := mihomonative.SettingsFromConfig(config)
	if mihomonative.SettingsEmpty(view) {
		return nil, nil
	}
	if err := settings.saveMihomoNative(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}
