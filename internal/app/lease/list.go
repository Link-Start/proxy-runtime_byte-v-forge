package lease

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
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

func (a *Application) List(ctx context.Context, options ListOptions) ([]*proxygatewayv1.ProxyDynamicLease, error) {
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

func (a *Application) list(ctx context.Context, options ListOptions) ([]*proxygatewayv1.ProxyDynamicLease, error) {
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

func ParseListOptions(query url.Values) (ListOptions, error) {
	options := DefaultListOptions()
	status := strings.TrimSpace(query.Get("status"))
	if status == "" && query.Get("include_inactive") == "true" {
		status = string(ListModeRecent)
	}
	if status != "" {
		mode, err := parseListMode(status)
		if err != nil {
			return ListOptions{}, err
		}
		options.Mode = mode
	}
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return ListOptions{}, fmt.Errorf("lease list limit must be a positive integer")
		}
		options.Limit = limit
	}
	return NormalizeListOptions(options), nil
}

func parseListMode(status string) (ListMode, error) {
	switch mode := ListMode(strings.TrimSpace(status)); mode {
	case ListModeActive, ListModeRecent, ListModeHistory:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported lease list status")
	}
}
