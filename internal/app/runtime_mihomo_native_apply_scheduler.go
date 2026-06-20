package app

import "github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"

type runtimeMihomoNativeApplyScheduler struct {
	exitCheckCache   *proxycheck.ExitCheckCache
	markApplyPending func()
	requestReconcile func()
}

func newRuntimeMihomoNativeApplyScheduler(runtime *Runtime) runtimeMihomoNativeApplyScheduler {
	if runtime == nil {
		return runtimeMihomoNativeApplyScheduler{}
	}
	return runtimeMihomoNativeApplyScheduler{
		exitCheckCache:   runtime.exitCheckCache,
		markApplyPending: runtime.markSettingsApplyPending,
		requestReconcile: runtime.requestReconcile,
	}
}

func (s runtimeMihomoNativeApplyScheduler) Schedule() {
	if s.exitCheckCache != nil {
		s.exitCheckCache.Clear()
	}
	if s.markApplyPending != nil {
		s.markApplyPending()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}
