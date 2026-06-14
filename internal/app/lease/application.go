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

type Application struct {
	repository Repository
}

func NewApplication(repository Repository) *Application {
	return &Application{repository: repository}
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
