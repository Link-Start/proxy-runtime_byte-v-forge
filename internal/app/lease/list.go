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

func (a *Application) List(ctx context.Context, options ListOptions) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	if a == nil || a.repository == nil {
		return nil, nil
	}
	options = NormalizeListOptions(options)
	startedAt := a.now()
	leases, err := a.list(ctx, options)
	if err != nil {
		a.warn("list proxy dynamic leases failed", "mode", options.Mode, "limit", options.Limit, "duration_ms", a.sinceMilliseconds(startedAt), "error_type", errorType(err))
		return nil, err
	}
	a.info("list proxy dynamic leases finished", "mode", options.Mode, "limit", options.Limit, "rows", len(leases), "duration_ms", a.sinceMilliseconds(startedAt))
	return leases, nil
}

func (a *Application) list(ctx context.Context, options ListOptions) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	switch options.Mode {
	case ListModeActive:
		return a.repository.ListActiveLeaseFacts(ctx, options.Limit)
	case ListModeRecent:
		return a.repository.ListRecentLeaseFacts(ctx, options.Limit)
	case ListModeHistory:
		return a.repository.ListHistoryLeaseFacts(ctx, options.Limit)
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
