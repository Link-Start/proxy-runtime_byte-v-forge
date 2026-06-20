package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type runtimeLeaseApplication struct {
	leases *leaseapp.Application
}

func newRuntimeLeaseApplication(deps leaseapp.Dependencies) runtimeLeaseApplication {
	return runtimeLeaseApplication{leases: leaseapp.NewApplication(deps)}
}

func runtimeLeaseDependencies(runtime *Runtime) leaseapp.Dependencies {
	if runtime == nil {
		return leaseapp.Dependencies{}
	}
	return leaseapp.Dependencies{
		Repository:  runtime.store,
		Coordinator: runtime.leaseCoordinator,
		Worker:      runtime.leaseCoordinator.workerProcessor(),
		Logger:      runtime.logger,
	}
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
