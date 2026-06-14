package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type runtimeLeaseApplication struct {
	logger *slog.Logger
	leases *leaseapp.Application
}

func newRuntimeLeaseApplication(runtime *Runtime) runtimeLeaseApplication {
	return runtimeLeaseApplication{logger: runtime.logger, leases: leaseapp.NewApplication(runtime.store, runtime.leaseCoordinator, runtime.leaseCoordinator)}
}

func (s *RuntimeService) ListProxyDynamicLeases(ctx context.Context, _ *proxyruntimev1.ListProxyDynamicLeasesRequest) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.listProxyDynamicLeases(ctx, leaseapp.DefaultListOptions())
}

func (s *RuntimeService) AcquireProxyLease(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.acquireProxyLease(ctx, nil, req)
}

func (s *RuntimeService) acquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.leases.AcquireProxyLease(ctx, advertisedProxyHost(httpReq), req)
}

func (s *RuntimeService) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	return s.leases.ReleaseProxyLease(ctx, req)
}

func (s *RuntimeService) listProxyDynamicLeases(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.leases.ListProxyDynamicLeaseFacts(ctx, options)
}

func (s *RuntimeService) getProxyDynamicLease(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leases.GetProxyDynamicLeaseFact(ctx, leaseID)
}

func (a runtimeLeaseApplication) GetProxyDynamicLeaseFact(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return a.leases.Get(ctx, leaseID)
}

func (a runtimeLeaseApplication) ListProxyDynamicLeaseFacts(ctx context.Context, options leaseapp.ListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	options = leaseapp.NormalizeListOptions(options)
	startedAt := time.Now()
	leases, err := a.leases.List(ctx, options)
	if err != nil {
		a.logger.Warn("list proxy dynamic leases failed", "mode", options.Mode, "limit", options.Limit, "duration_ms", time.Since(startedAt).Milliseconds(), "error", err)
		return nil, err
	}
	a.logger.Info("list proxy dynamic leases finished", "mode", options.Mode, "limit", options.Limit, "rows", len(leases), "duration_ms", time.Since(startedAt).Milliseconds())
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
