package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (r *Runtime) listenerReservedLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	active, err := r.store.ListActiveLeaseFacts(ctx, leaseapp.MaxListLimit)
	if err != nil {
		return nil, err
	}
	cleanupPending, err := r.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		return nil, err
	}
	return listenerReservedLeaseFacts(active, cleanupPending), nil
}

func listenerReservedLeaseFacts(active []*proxyruntimev1.ProxyDynamicLease, cleanupPending []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.ProxyDynamicLease {
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(active)+len(cleanupPending))
	seen := map[string]struct{}{}
	appendReserved := func(lease *proxyruntimev1.ProxyDynamicLease) {
		if lease.GetListener() == nil {
			return
		}
		leaseID := strings.TrimSpace(lease.GetLeaseId())
		if leaseID != "" {
			if _, exists := seen[leaseID]; exists {
				return
			}
			seen[leaseID] = struct{}{}
		}
		out = append(out, lease)
	}
	for _, lease := range active {
		if leaseapp.HasActiveStatus(lease) {
			appendReserved(lease)
		}
	}
	for _, lease := range cleanupPending {
		if leaseapp.RouteCleanupPending(lease) {
			appendReserved(lease)
		}
	}
	return out
}
