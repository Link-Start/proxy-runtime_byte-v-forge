package lease

import (
	"context"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a *Application) Acquire(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	if a == nil || a.coordinator == nil {
		return nil, fmt.Errorf("lease coordinator is required")
	}
	lease, err := a.coordinator.Acquire(ctx, advertisedHost, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.AcquireProxyLeaseResponse{Lease: lease, Egress: lease.GetEgress(), SelectionPlan: lease.GetSelectionPlan()}, nil
}

func (a *Application) Release(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	if a == nil || a.coordinator == nil {
		return nil, fmt.Errorf("lease coordinator is required")
	}
	lease, err := a.coordinator.Release(ctx, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ReleaseProxyLeaseResponse{Lease: lease}, nil
}

func (a *Application) now() time.Time {
	if a != nil && a.clock != nil {
		return a.clock.Now()
	}
	return time.Now()
}

func (a *Application) sinceMilliseconds(startedAt time.Time) int64 {
	return a.now().Sub(startedAt).Milliseconds()
}

func (a *Application) info(message string, args ...any) {
	if a != nil && a.logger != nil {
		a.logger.Info(message, args...)
	}
}

func (a *Application) warn(message string, args ...any) {
	if a != nil && a.logger != nil {
		a.logger.Warn(message, args...)
	}
}
