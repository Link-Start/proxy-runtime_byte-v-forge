package lease

import (
	"context"
	"fmt"
	"reflect"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (a *Application) Acquire(ctx context.Context, advertisedHost string, req *proxygatewayv1.AcquireProxyLeaseRequest) (*proxygatewayv1.AcquireProxyLeaseResponse, error) {
	if a == nil || a.coordinator == nil {
		return nil, fmt.Errorf("lease coordinator is required")
	}
	startedAt := a.now()
	lease, err := a.coordinator.Acquire(ctx, advertisedHost, req)
	if err != nil {
		a.warn("acquire proxy dynamic lease failed", "account_id", req.GetAccountId(), "purpose", req.GetPurpose(), "duration_ms", a.sinceMilliseconds(startedAt), "error_type", errorType(err))
		return nil, err
	}
	a.info("acquire proxy dynamic lease finished", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "purpose", lease.GetPurpose(), "provider_account_key", lease.GetProviderAccountId(), "duration_ms", a.sinceMilliseconds(startedAt))
	return &proxygatewayv1.AcquireProxyLeaseResponse{Lease: lease, Egress: lease.GetEgress(), SelectionPlan: lease.GetSelectionPlan()}, nil
}

func (a *Application) Release(ctx context.Context, req *proxygatewayv1.ReleaseProxyLeaseRequest) (*proxygatewayv1.ReleaseProxyLeaseResponse, error) {
	if a == nil || a.coordinator == nil {
		return nil, fmt.Errorf("lease coordinator is required")
	}
	startedAt := a.now()
	lease, err := a.coordinator.Release(ctx, req)
	if err != nil {
		a.warn("release proxy dynamic lease failed", "lease_id", req.GetLeaseId(), "account_id", req.GetAccountId(), "purpose", req.GetPurpose(), "duration_ms", a.sinceMilliseconds(startedAt), "error_type", errorType(err))
		return nil, err
	}
	a.info("release proxy dynamic lease finished", "lease_id", lease.GetLeaseId(), "account_id", lease.GetAccountId(), "purpose", lease.GetPurpose(), "provider_account_key", lease.GetProviderAccountId(), "duration_ms", a.sinceMilliseconds(startedAt))
	return &proxygatewayv1.ReleaseProxyLeaseResponse{Lease: lease}, nil
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

func errorType(err error) string {
	if err == nil {
		return ""
	}
	return reflect.TypeOf(err).String()
}
