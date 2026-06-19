package clock

import "time"

// Clock is the injectable time source shared across the runtime. Business
// logic depends on it instead of calling time.Now() directly so that expiry,
// TTL and timestamp decisions stay deterministic and replayable.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock backed by the wall clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}
