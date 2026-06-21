package lease

import (
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

const (
	RouteCleanupPendingLabel    = "route_cleanup_pending"
	ProviderCleanupPendingLabel = "provider_cleanup_pending"
	CleanupFinalStatusLabel     = "cleanup_final_status"
)

func HasLeaseID(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return lease != nil && strings.TrimSpace(lease.GetLeaseId()) != ""
}

func ActiveAt(lease *proxygatewayv1.ProxyDynamicLease, now time.Time) bool {
	if !HasActiveStatus(lease) {
		return false
	}
	return lease.GetExpiresAt() == nil || now.Before(lease.GetExpiresAt().AsTime())
}

func NeedsExpiryCleanup(lease *proxygatewayv1.ProxyDynamicLease, now time.Time) bool {
	return HasActiveStatus(lease) && !ActiveAt(lease, now)
}

func HasActiveStatus(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE)
}

func HasReleasedStatus(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED)
}

func HasExpiredStatus(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_EXPIRED)
}

func HasFailedStatus(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return hasStatus(lease, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED)
}

func HasSuccessfulAttemptStatus(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return HasActiveStatus(lease) || HasExpiredStatus(lease) || HasReleasedStatus(lease)
}

func hasStatus(lease *proxygatewayv1.ProxyDynamicLease, status proxygatewayv1.ProxyDynamicLeaseStatus) bool {
	return lease != nil && lease.GetStatus() == status
}

func CleanupPending(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return RouteCleanupPending(lease) || ProviderCleanupPending(lease)
}

func RouteCleanupPending(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return CleanupLabel(lease, RouteCleanupPendingLabel) == "true"
}

func ProviderCleanupPending(lease *proxygatewayv1.ProxyDynamicLease) bool {
	return CleanupLabel(lease, ProviderCleanupPendingLabel) == "true"
}

func CleanupFinalStatus(lease *proxygatewayv1.ProxyDynamicLease) string {
	return CleanupLabel(lease, CleanupFinalStatusLabel)
}

func CleanupLabel(lease *proxygatewayv1.ProxyDynamicLease, key string) string {
	if lease == nil || lease.GetSession() == nil {
		return ""
	}
	return strings.TrimSpace(lease.GetSession().GetLabels()[key])
}
