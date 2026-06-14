package lease

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ExistingActiveLeaseDecision int

const (
	ExistingActiveLeaseIgnore ExistingActiveLeaseDecision = iota
	ExistingActiveLeaseReuse
	ExistingActiveLeaseReplace
)

func DecideExistingActiveLease(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease, now time.Time, playgroundAccountID string, playgroundUsername string) ExistingActiveLeaseDecision {
	if !ActiveAt(lease, now) {
		return ExistingActiveLeaseIgnore
	}
	if !req.GetForceNew() && !PlaygroundLeaseNeedsReplacement(req, lease, playgroundAccountID, playgroundUsername) {
		return ExistingActiveLeaseReuse
	}
	return ExistingActiveLeaseReplace
}
