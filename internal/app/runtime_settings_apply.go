package app

import (
	"context"
	"time"
)

const runtimeSettingsConnectionCleanupTimeout = 10 * time.Second

func (a runtimeSettingsApplication) scheduleRuntimeSettingsApply(changedUsernames []string) {
	if a.scheduleApply == nil {
		return
	}
	a.scheduleApply(changedUsernames)
}

func (r *Runtime) scheduleRuntimeSettingsApply(changedUsernames []string) {
	r.resetIPFraudChecker()
	r.geoCache.clear()
	r.exitCheckCache.clear()
	r.requestReconcile()
	if len(changedUsernames) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), runtimeSettingsConnectionCleanupTimeout)
		defer cancel()
		r.closeMihomoInUserConnections(ctx, changedUsernames)
	}()
}
