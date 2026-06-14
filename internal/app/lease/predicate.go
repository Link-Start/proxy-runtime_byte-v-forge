package lease

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	RouteCleanupPendingLabel    = "route_cleanup_pending"
	ProviderCleanupPendingLabel = "provider_cleanup_pending"
	CleanupFinalStatusLabel     = "cleanup_final_status"
)

func ActiveAt(lease *proxyruntimev1.ProxyDynamicLease, now time.Time) bool {
	if lease == nil || lease.GetStatus() != proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE {
		return false
	}
	return lease.GetExpiresAt() == nil || now.Before(lease.GetExpiresAt().AsTime())
}

func CleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return RouteCleanupPending(lease) || ProviderCleanupPending(lease)
}

func RouteCleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return CleanupLabel(lease, RouteCleanupPendingLabel) == "true"
}

func ProviderCleanupPending(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return CleanupLabel(lease, ProviderCleanupPendingLabel) == "true"
}

func CleanupFinalStatus(lease *proxyruntimev1.ProxyDynamicLease) string {
	return CleanupLabel(lease, CleanupFinalStatusLabel)
}

func CleanupLabel(lease *proxyruntimev1.ProxyDynamicLease, key string) string {
	if lease == nil || lease.GetSession() == nil {
		return ""
	}
	return strings.TrimSpace(lease.GetSession().GetLabels()[key])
}
