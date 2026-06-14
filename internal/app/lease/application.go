package lease

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ListMode string

const (
	ListModeActive  ListMode = "active"
	ListModeRecent  ListMode = "recent"
	ListModeHistory ListMode = "history"

	DefaultListLimit = 50
	MaxListLimit     = 200
)

type ListOptions struct {
	Mode  ListMode
	Limit int
}

type Repository interface {
	ListActiveLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ListRecentLeaseFacts(context.Context, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
}

type Coordinator interface {
	Acquire(context.Context, string, *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error)
	Release(context.Context, *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error)
}

type Application struct {
	repository  Repository
	coordinator Coordinator
}

func NewApplication(repository Repository, coordinator Coordinator) *Application {
	return &Application{repository: repository, coordinator: coordinator}
}

func (a *Application) List(ctx context.Context, options ListOptions) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	if a == nil || a.repository == nil {
		return nil, nil
	}
	options = NormalizeListOptions(options)
	switch options.Mode {
	case ListModeActive:
		return a.repository.ListActiveLeaseFacts(ctx, options.Limit)
	case ListModeRecent, ListModeHistory:
		return a.repository.ListRecentLeaseFacts(ctx, options.Limit)
	default:
		return nil, fmt.Errorf("unsupported lease list status %q", options.Mode)
	}
}

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

func DefaultListOptions() ListOptions {
	return ListOptions{Mode: ListModeActive, Limit: DefaultListLimit}
}

func NormalizeListOptions(options ListOptions) ListOptions {
	if options.Mode == "" {
		options.Mode = ListModeActive
	}
	options.Limit = NormalizeListLimit(options.Limit)
	return options
}

func NormalizeListLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}
