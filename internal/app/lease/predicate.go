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
	if !HasActiveStatus(lease) {
		return false
	}
	return lease.GetExpiresAt() == nil || now.Before(lease.GetExpiresAt().AsTime())
}

func HasActiveStatus(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE)
}

func HasReleasedStatus(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED)
}

func HasExpiredStatus(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED)
}

func HasFailedStatus(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED)
}

func HasSuccessfulAttemptStatus(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return HasActiveStatus(lease) || HasExpiredStatus(lease) || HasReleasedStatus(lease)
}

func hasStatus(lease *proxyruntimev1.ProxyDynamicLease, status proxyruntimev1.ProxyDynamicLeaseStatus) bool {
	return lease != nil && lease.GetStatus() == status
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
