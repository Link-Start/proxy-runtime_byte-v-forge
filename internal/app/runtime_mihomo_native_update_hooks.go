package app

func runMihomoNativeUpdateAfterApply(callback func()) {
	if callback != nil {
		callback()
	}
}
