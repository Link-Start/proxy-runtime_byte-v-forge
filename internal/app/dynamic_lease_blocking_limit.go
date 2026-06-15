package app

const (
	defaultBlockingLeaseFactLimit = 20
	maxBlockingLeaseFactLimit     = 100
)

func normalizeBlockingLeaseFactLimit(limit int) int {
	if limit <= 0 {
		return defaultBlockingLeaseFactLimit
	}
	if limit > maxBlockingLeaseFactLimit {
		return maxBlockingLeaseFactLimit
	}
	return limit
}
