package app

import (
	"context"
	"net/http"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type runtimeLeaseApplication struct {
	leases *leaseapp.Application
}

func newRuntimeLeaseApplication(deps leaseapp.Dependencies) runtimeLeaseApplication {
	return runtimeLeaseApplication{leases: leaseapp.NewApplication(deps)}
}

func (s *RuntimeService) ListProxyDynamicLeases(ctx context.Context, _ *proxyruntimev1.ListProxyDynamicLeasesRequest) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.listProxyDynamicLeases(ctx, leaseapp.DefaultListOptions())
}

func (s *RuntimeService) AcquireProxyLease(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.acquireProxyLease(ctx, nil, req)
}

func (s *RuntimeService) acquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	startedAt := time.Now()
	response, err := s.leases.AcquireProxyLease(ctx, advertisedProxyHost(httpReq), req)
	s.observeServiceOperation(runtimeMetricLeaseAcquire, startedAt, err)
	return response, err
}

func (s *RuntimeService) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	startedAt := time.Now()
	response, err := s.leases.ReleaseProxyLease(ctx, req)
	s.observeServiceOperation(runtimeMetricLeaseRelease, startedAt, err)
	return response, err
}

func (s *RuntimeService) listProxyDynamicLeases(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	startedAt := time.Now()
	response, err := s.leases.ListProxyDynamicLeaseFacts(ctx, options)
	s.observeServiceOperation(runtimeMetricLeaseList, startedAt, err)
	return response, err
}

func (s *RuntimeService) getProxyDynamicLease(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leases.GetProxyDynamicLeaseFact(ctx, leaseID)
}

func (a runtimeLeaseApplication) GetProxyDynamicLeaseFact(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return a.leases.Get(ctx, leaseID)
}

func (a runtimeLeaseApplication) ListProxyDynamicLeaseFacts(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	leases, err := a.leases.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyDynamicLeasesResponse{Leases: leases}, nil
}

func (a runtimeLeaseApplication) AcquireProxyLease(ctx context.Context, advertisedHost string, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return a.leases.Acquire(ctx, advertisedHost, req)
}

func (a runtimeLeaseApplication) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	return a.leases.Release(ctx, req)
}

func (a runtimeLeaseApplication) RestoreActiveLeases(ctx context.Context) error {
	return a.leases.RestoreActive(ctx)
}

func (a runtimeLeaseApplication) ExpireDueLeaseFacts(ctx context.Context) error {
	return a.leases.ExpireDue(ctx)
}

func (a runtimeLeaseApplication) CleanupPendingLeaseFacts(ctx context.Context) error {
	return a.leases.CleanupPending(ctx)
}

func (a runtimeLeaseApplication) CleanupPendingLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	return a.leases.Cleanup(ctx, lease)
}

func (s *RuntimeService) observeServiceOperation(operation string, startedAt time.Time, err error) {
	if s != nil && s.metrics != nil {
		s.metrics.Observe(operation, startedAt, err)
	}
}
