package app

import (
	"context"
	"time"
)

const runtimeSettingsConnectionCleanupTimeout = 10 * time.Second

func (a runtimeSettingsApplication) scheduleRuntimeSettingsApply(changedUsernames []string) {
	a.runtime.resetIPFraudChecker()
	a.runtime.geoCache.clear()
	a.runtime.exitCheckCache.clear()
	a.runtime.requestReconcile()
	if len(changedUsernames) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), runtimeSettingsConnectionCleanupTimeout)
		defer cancel()
		a.runtime.closeMihomoInUserConnections(ctx, changedUsernames)
	}()
}
