package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func ReservedListenerLeaseFacts(active []*proxyruntimev1.ProxyDynamicLease, cleanupPending []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.ProxyDynamicLease {
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
		if HasActiveStatus(lease) {
			appendReserved(lease)
		}
	}
	for _, lease := range cleanupPending {
		if RouteCleanupPending(lease) {
			appendReserved(lease)
		}
	}
	return out
}
