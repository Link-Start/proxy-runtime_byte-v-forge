package lease

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type BatchPredicate func(*proxygatewayv1.ProxyDynamicLease) bool

type BatchProcessor func(context.Context, *proxygatewayv1.ProxyDynamicLease) error

type BatchErrorObserver func(*proxygatewayv1.ProxyDynamicLease, error)

type BatchInput struct {
	Leases      []*proxygatewayv1.ProxyDynamicLease
	Timeout     time.Duration
	ErrorPrefix string
	ShouldRun   BatchPredicate
	Process     BatchProcessor
	Observe     BatchErrorObserver
}

func ProcessLeaseBatch(ctx context.Context, input BatchInput) error {
	if input.Process == nil {
		return nil
	}
	batchErrors := make([]error, 0)
	for _, lease := range input.Leases {
		if input.ShouldRun != nil && !input.ShouldRun(lease) {
			continue
		}
		if err := ctx.Err(); err != nil {
			batchErrors = append(batchErrors, err)
			break
		}
		attemptCtx, cancel := batchAttemptContext(ctx, input.Timeout)
		err := input.Process(attemptCtx, lease)
		cancel()
		if err != nil {
			if input.Observe != nil {
				input.Observe(lease, err)
			}
			batchErrors = append(batchErrors, batchLeaseError(input.ErrorPrefix, lease, err))
		}
	}
	return errors.Join(batchErrors...)
}

func batchAttemptContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func batchLeaseError(prefix string, lease *proxygatewayv1.ProxyDynamicLease, err error) error {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "process lease"
	}
	return fmt.Errorf("%s %q: %w", prefix, lease.GetLeaseId(), err)
}
