package lease

import (
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrAcquireAttemptRunnerRequired = errors.New("acquire attempt runner is required")

type AcquireAttemptRunner func(attempt int) (*proxyruntimev1.ProxyDynamicLease, error)

type AcquireAttemptRetryPolicy func(error) bool

type AcquireAttemptFailureObserver func(attempt int, err error)

func RunAcquireAttempts(req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPolicy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, run AcquireAttemptRunner, retry AcquireAttemptRetryPolicy, observe AcquireAttemptFailureObserver) (*proxyruntimev1.ProxyDynamicLease, error) {
	if run == nil {
		return nil, ErrAcquireAttemptRunnerRequired
	}
	var lastErr error
	for attempt := 1; attempt <= DynamicIPSelectionMaxAttempts(selectionPolicy); attempt++ {
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
