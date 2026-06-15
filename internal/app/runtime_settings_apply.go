package app

import (
	"context"
	"time"
)

const runtimeSettingsConnectionCleanupTimeout = 10 * time.Second

type runtimeSettingsApplyScheduler struct {
	resetIPFraudChecker   func()
	geoCache              *ipGeoCache
	exitCheckCache        *proxyExitCheckCache
	markApplyPending      func()
	requestReconcile      func()
	closeInUserConnection leaseConnectionCleanupFunc
}

func newRuntimeSettingsApplyScheduler(runtime *Runtime) runtimeSettingsApplyScheduler {
	if runtime == nil {
		return runtimeSettingsApplyScheduler{}
	}
	return runtimeSettingsApplyScheduler{
		resetIPFraudChecker:   runtime.resetIPFraudChecker,
		geoCache:              &runtime.geoCache,
		exitCheckCache:        &runtime.exitCheckCache,
		markApplyPending:      runtime.markSettingsApplyPending,
		requestReconcile:      runtime.requestReconcile,
		closeInUserConnection: runtime.closeMihomoInUserConnections,
	}
}

func (s runtimeSettingsApplyScheduler) Schedule(changedUsernames []string) {
	s.clearDerivedState()
	s.requestApply()
	if len(changedUsernames) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), runtimeSettingsConnectionCleanupTimeout)
		defer cancel()
		s.closeInUserConnections(ctx, changedUsernames)
	}()
}

func (s runtimeSettingsApplyScheduler) clearDerivedState() {
	if s.resetIPFraudChecker != nil {
		s.resetIPFraudChecker()
	}
	if s.geoCache != nil {
		s.geoCache.clear()
	}
	if s.exitCheckCache != nil {
		s.exitCheckCache.clear()
	}
}

func (s runtimeSettingsApplyScheduler) requestApply() {
	if s.markApplyPending != nil {
		s.markApplyPending()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}

func (s runtimeSettingsApplyScheduler) closeInUserConnections(ctx context.Context, changedUsernames []string) {
	if s.closeInUserConnection != nil {
		s.closeInUserConnection(ctx, changedUsernames)
	}
}
