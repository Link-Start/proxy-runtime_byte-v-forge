package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"
)

func loadMihomoNativeUpdateCurrent(ctx context.Context, repository mihomoNativeUpdateRepository) (mihomonative.ConfigFile, error) {
	currentView, err := repository.loadMihomoNative(ctx)
	if err != nil {
		return mihomonative.ConfigFile{}, internalError("load mihomo native settings", err)
	}
	current, err := mihomonative.ConfigFromSettings(currentView)
	if err != nil {
		return mihomonative.ConfigFile{}, internalError("load mihomo native settings", err)
	}
	return current, nil
}
