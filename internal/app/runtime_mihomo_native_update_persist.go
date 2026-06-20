package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func persistMihomoNativeUpdatePlan(ctx context.Context, deps mihomoNativeUpdateDependencies, plan mihomonative.UpdatePlan) error {
	if err := deps.Repository.saveMihomoNative(ctx, mihomonative.SettingsFromConfig(plan.Config)); err != nil {
		return appcore.InternalError("save mihomo native settings", err)
	}
	if err := mihomonative.SaveConfig(deps.ConfigDir, plan.Config); err != nil {
		return appcore.InternalError("save mihomo native config", err)
	}
	if _, err := deps.Repository.replaceMihomoResourceRefs(ctx, plan.ResourceReplacements); err != nil {
		return appcore.InternalError("update mihomo native resource references", err)
	}
	return nil
}
