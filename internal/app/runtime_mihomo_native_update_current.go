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
