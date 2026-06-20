package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
	mihomoapp "github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative/application"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/settings/adapter/persistence"
)

type runtimeSettingsMihomoNativeAdapter struct {
	repository *persistence.Store
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
		return mihomonative.NormalizeSettings(nil), nil
	}
	return a.repository.LoadMihomoNative(ctx)
}

func (a runtimeSettingsMihomoNativeAdapter) Update(ctx context.Context, config *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if a.repository == nil {
		return nil, appcore.InternalError("mihomo native settings repository is required", nil)
	}
	return mihomoapp.Update(ctx, mihomoapp.UpdateDependencies{
		ConfigDir:    a.configDir,
		LoadSettings: a.repository.LoadMihomoNative,
		SaveSettings: a.repository.SaveMihomoNative,
		ReplaceRefs:  a.repository.ReplaceMihomoResourceRefs,
		AfterApply:   a.apply.Schedule,
	}, config)
}
