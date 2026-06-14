package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func updateMihomoNativeSettings(ctx context.Context, runtime *Runtime, view *proxyruntimev1.ProxyRuntimeMihomoNativeConfig) (*proxyruntimev1.ProxyRuntimeMihomoNativeConfig, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if runtime == nil {
		return nil, internalError("runtime is required", nil)
	}
	currentView, err := runtime.settings.loadMihomoNative(ctx)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	current, err := mihomoNativeConfigFileFromSettings(currentView)
	if err != nil {
		return nil, internalError("load mihomo native settings", err)
	}
	plan, err := buildMihomoNativeUpdatePlan(current, view)
	if err != nil {
		return nil, err
	}
	if err := runtime.settings.saveMihomoNative(ctx, mihomoNativeSettingsFromConfig(plan.Config)); err != nil {
		return nil, internalError("save mihomo native settings", err)
	}
	if err := saveMihomoNativeConfig(runtime, plan.Config); err != nil {
		return nil, internalError("save mihomo native config", err)
	}
	_, err = runtime.settings.replaceMihomoResourceRefs(ctx, plan.ResourceReplacements)
	if err != nil {
		return nil, internalError("update mihomo native resource references", err)
	}
	runtime.exitCheckCache.clear()
	runtime.requestReconcile()
	return mihomoNativeSettings(ctx, runtime)
}
