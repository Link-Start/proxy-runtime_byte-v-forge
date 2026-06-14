package app

import (
	"context"
	"net/http"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeLeaseApplication struct {
	runtime *Runtime
	leases  leaseCoordinator
}

type leaseListMode string

const (
	leaseListModeActive  leaseListMode = "active"
	leaseListModeRecent  leaseListMode = "recent"
	leaseListModeHistory leaseListMode = "history"

	defaultLeaseListLimit = 50
	maxLeaseListLimit     = 200
)

type leaseListOptions struct {
	mode  leaseListMode
	limit int
}

func newRuntimeLeaseApplication(runtime *Runtime) runtimeLeaseApplication {
	return runtimeLeaseApplication{runtime: runtime, leases: runtime.leaseCoordinator}
}

func (s *RuntimeService) ListProxyDynamicLeases(ctx context.Context, _ *proxyruntimev1.ListProxyDynamicLeasesRequest) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.listProxyDynamicLeases(ctx, defaultLeaseListOptions())
}

func (s *RuntimeService) AcquireProxyLease(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.acquireProxyLease(ctx, nil, req)
}

func (s *RuntimeService) acquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	return s.leases.AcquireProxyLease(ctx, httpReq, req)
}

func (s *RuntimeService) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	return s.leases.ReleaseProxyLease(ctx, req)
}

func (s *RuntimeService) listProxyDynamicLeases(ctx context.Context, options leaseListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	return s.leases.ListProxyDynamicLeaseFacts(ctx, options)
}

func (a runtimeLeaseApplication) ListProxyDynamicLeaseFacts(ctx context.Context, options leaseListOptions) (*proxyruntimev1.ListProxyDynamicLeasesResponse, error) {
	options = normalizeLeaseListOptions(options)
	var (
		leases []*proxyruntimev1.ProxyDynamicLease
		err    error
	)
	switch options.mode {
	case leaseListModeActive:
		leases, err = a.runtime.store.ListActiveLeaseFacts(ctx, options.limit)
	case leaseListModeRecent, leaseListModeHistory:
		leases, err = a.runtime.store.ListRecentLeaseFacts(ctx, options.limit)
	default:
		return nil, invalidArgument("unsupported lease list status", nil)
	}
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyDynamicLeasesResponse{Leases: leases}, nil
}

func (a runtimeLeaseApplication) AcquireProxyLease(ctx context.Context, httpReq *http.Request, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.AcquireProxyLeaseResponse, error) {
	lease, err := a.leases.acquireLease(ctx, httpReq, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.AcquireProxyLeaseResponse{Lease: lease, Egress: lease.GetEgress(), SelectionPlan: lease.GetSelectionPlan()}, nil
}

func (a runtimeLeaseApplication) ReleaseProxyLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error) {
	lease, err := a.leases.releaseLease(ctx, req)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ReleaseProxyLeaseResponse{Lease: lease}, nil
}

func defaultLeaseListOptions() leaseListOptions {
	return leaseListOptions{mode: leaseListModeActive, limit: defaultLeaseListLimit}
}

func normalizeLeaseListOptions(options leaseListOptions) leaseListOptions {
	if options.mode == "" {
		options.mode = leaseListModeActive
	}
	options.limit = normalizeLeaseListLimit(options.limit)
	return options
}

func normalizeLeaseListLimit(limit int) int {
	if limit <= 0 {
		return defaultLeaseListLimit
	}
	if limit > maxLeaseListLimit {
		return maxLeaseListLimit
	}
	return limit
}
