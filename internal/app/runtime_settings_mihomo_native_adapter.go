package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsMihomoNativeAdapter struct {
	repository mihomoNativeUpdateRepository
	configDir  string
	apply      runtimeMihomoNativeApplyScheduler
}

func newRuntimeSettingsMihomoNativeAdapter(runtime *Runtime) runtimeSettingsMihomoNativeAdapter {
	if runtime == nil {
		return runtimeSettingsMihomoNativeAdapter{}
	}
	return runtimeSettingsMihomoNativeAdapter{
		repository: runtime.settings,
		configDir:  runtime.cfg.Mihomo.ConfigDir,
		apply:      newRuntimeMihomoNativeApplyScheduler(runtime),
	}
}

func (a runtimeSettingsMihomoNativeAdapter) Load(ctx context.Context) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.repository == nil {
		return normalizeMihomoNativeSettings(nil), nil
	}
	return a.repository.loadMihomoNative(ctx)
}

func (a runtimeSettingsMihomoNativeAdapter) Update(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	return updateMihomoNativeSettings(ctx, mihomoNativeUpdateDependencies{
		Repository: a.repository,
		ConfigDir:  a.configDir,
		AfterApply: a.apply.Schedule,
	}, config)
}
