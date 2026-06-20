package app

func defaultProviderWeight(index int) uint32 {
	if index < 0 {
		return 100
	}
	if index > 9 {
		return 10
	}
	return uint32(100 - index*10)
}
