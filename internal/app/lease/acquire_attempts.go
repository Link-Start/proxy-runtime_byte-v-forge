package lease

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAcquireAttemptRunnerRequired = errors.New("acquire attempt runner is required")

type AcquireAttemptRunner func(attempt int) (*proxyruntimev1.ProxyDynamicLease, error)

type AcquireAttemptRetryPolicy func(error) bool

type AcquireAttemptFailureObserver func(attempt int, err error)

const (
	acquireRetryBaseBackoff = 100 * time.Millisecond
	acquireRetryMaxBackoff  = 2 * time.Second
)

func RunAcquireAttempts(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPolicy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, run AcquireAttemptRunner, retry AcquireAttemptRetryPolicy, observe AcquireAttemptFailureObserver) (*proxyruntimev1.ProxyDynamicLease, error) {
	if run == nil {
		return nil, ErrAcquireAttemptRunnerRequired
	}
	maxAttempts := DynamicIPSelectionMaxAttempts(selectionPolicy)
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 1 {
			if err := sleepAcquireBackoff(ctx, attempt); err != nil {
				return nil, err
			}
		}
		SetAttemptLabel(req, attempt)
		lease, err := run(attempt)
		if err == nil {
			return lease, nil
		}
		lastErr = err
		if retry == nil || !retry(err) {
			return nil, err
		}
		if observe != nil {
			observe(attempt, err)
		}
	}
	return nil, lastErr
}

func sleepAcquireBackoff(ctx context.Context, attempt int) error {
	shift := attempt - 2
	if shift < 0 {
		shift = 0
	}
	if shift > 16 {
		shift = 16
	}
	backoff := acquireRetryBaseBackoff << shift
	if backoff <= 0 || backoff > acquireRetryMaxBackoff {
		backoff = acquireRetryMaxBackoff
	}
	timer := time.NewTimer(rand.N(backoff))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
