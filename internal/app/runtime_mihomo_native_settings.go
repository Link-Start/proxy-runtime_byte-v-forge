package app

import (
	"context"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func mihomoNativeSettings(ctx context.Context, runtime *Runtime) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if runtime == nil || runtime.settings == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return runtime.settings.loadMihomoNative(ctx)
}
func mihomoNativeSettingsEmpty(view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) bool {
	return view == nil || len(view.GetFixedProxies()) == 0 && len(view.GetSubscriptions()) == 0
}
