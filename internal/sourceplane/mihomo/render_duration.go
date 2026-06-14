package mihomo

import "time"

func defaultExpectedStatus(value uint32) uint32 {
	if value == 0 {
		return 204
	}
	return value
}

func seconds(value time.Duration, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return int(value / time.Second)
}

func milliseconds(value time.Duration, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return int(value / time.Millisecond)
}

func secondsDuration(value time.Duration, fallback int) int { return seconds(value, fallback) }

func millisecondsDuration(value time.Duration, fallback int) int {
	return milliseconds(value, fallback)
}
