package app

import "context"

func loadMihomoNativeUpdateCurrent(ctx context.Context, repository mihomoNativeUpdateRepository) (mihomoNativeConfigFile, error) {
	currentView, err := repository.loadMihomoNative(ctx)
	if err != nil {
		return mihomoNativeConfigFile{}, internalError("load mihomo native settings", err)
	}
	current, err := mihomoNativeConfigFileFromSettings(currentView)
	if err != nil {
		return mihomoNativeConfigFile{}, internalError("load mihomo native settings", err)
	}
	return current, nil
}

func persistMihomoNativeUpdatePlan(ctx context.Context, deps mihomoNativeUpdateDependencies, plan mihomoNativeUpdatePlan) error {
	if err := deps.Repository.saveMihomoNative(ctx, mihomoNativeSettingsFromConfig(plan.Config)); err != nil {
		return internalError("save mihomo native settings", err)
	}
	if err := saveMihomoNativeConfig(deps.ConfigDir, plan.Config); err != nil {
		return internalError("save mihomo native config", err)
	}
	if _, err := deps.Repository.replaceMihomoResourceRefs(ctx, plan.ResourceReplacements); err != nil {
		return internalError("update mihomo native resource references", err)
	}
	return nil
}
