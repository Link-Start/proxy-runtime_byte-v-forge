package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func loadMihomoNativeUpdateCurrent(ctx context.Context, repository mihomoNativeUpdateRepository) (mihomonative.ConfigFile, error) {
	currentView, err := repository.loadMihomoNative(ctx)
	if err != nil {
		return mihomonative.ConfigFile{}, appcore.InternalError("load mihomo native settings", err)
	}
	current, err := mihomonative.ConfigFromSettings(currentView)
	if err != nil {
		return mihomonative.ConfigFile{}, appcore.InternalError("load mihomo native settings", err)
	}
	return current, nil
}
