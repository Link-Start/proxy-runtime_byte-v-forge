package app

import (
	"sort"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func filterLeaseFacts(in []*proxyruntimev1.ProxyDynamicLease, keep func(*proxyruntimev1.ProxyDynamicLease) bool) []*proxyruntimev1.ProxyDynamicLease {
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for _, lease := range in {
		if keep(lease) {
			out = append(out, lease)
		}
	}
	return out
}

func sortLeaseFactsByAcquiredDesc(leases []*proxyruntimev1.ProxyDynamicLease) {
	sort.SliceStable(leases, func(left, right int) bool {
		return leaseSortTime(leases[left]).After(leaseSortTime(leases[right]))
	})
}

func leaseSortTime(lease *proxyruntimev1.ProxyDynamicLease) time.Time {
	if lease.GetAcquiredAt() != nil {
		return lease.GetAcquiredAt().AsTime()
	}
	if lease.GetExpiresAt() != nil {
		return lease.GetExpiresAt().AsTime()
	}
	return time.Time{}
}
