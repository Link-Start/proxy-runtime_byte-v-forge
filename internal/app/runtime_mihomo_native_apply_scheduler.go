package app

type runtimeMihomoNativeApplyScheduler struct {
	exitCheckCache   *proxyExitCheckCache
	markApplyPending func()
	requestReconcile func()
}

func newRuntimeMihomoNativeApplyScheduler(runtime *Runtime) runtimeMihomoNativeApplyScheduler {
	if runtime == nil {
		return runtimeMihomoNativeApplyScheduler{}
	}
	return runtimeMihomoNativeApplyScheduler{
		exitCheckCache:   &runtime.exitCheckCache,
		markApplyPending: runtime.markSettingsApplyPending,
		requestReconcile: runtime.requestReconcile,
	}
}

func (s runtimeMihomoNativeApplyScheduler) Schedule() {
	if s.exitCheckCache != nil {
		s.exitCheckCache.clear()
	}
	if s.markApplyPending != nil {
		s.markApplyPending()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}
