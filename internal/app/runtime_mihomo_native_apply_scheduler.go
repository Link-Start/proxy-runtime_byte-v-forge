package app

type runtimeMihomoNativeApplyScheduler struct {
	exitCheckCache   *proxyExitCheckCache
	requestReconcile func()
}

func newRuntimeMihomoNativeApplyScheduler(runtime *Runtime) runtimeMihomoNativeApplyScheduler {
	if runtime == nil {
		return runtimeMihomoNativeApplyScheduler{}
	}
	return runtimeMihomoNativeApplyScheduler{
		exitCheckCache:   &runtime.exitCheckCache,
		requestReconcile: runtime.requestReconcile,
	}
}

func (s runtimeMihomoNativeApplyScheduler) Schedule() {
	if s.exitCheckCache != nil {
		s.exitCheckCache.clear()
	}
	if s.requestReconcile != nil {
		s.requestReconcile()
	}
}
