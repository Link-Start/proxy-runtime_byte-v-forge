package kernel

const (
	DefaultBlockingLeaseFactLimit = 20
	maxBlockingLeaseFactLimit     = 100
)

func NormalizeBlockingLeaseFactLimit(limit int) int {
	if limit <= 0 {
		return DefaultBlockingLeaseFactLimit
	}
	if limit > maxBlockingLeaseFactLimit {
		return maxBlockingLeaseFactLimit
	}
	return limit
}
