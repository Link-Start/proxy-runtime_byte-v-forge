package lease

import (
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func ReuseExistingActiveLease(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease, now time.Time, playgroundAccountID string, playgroundUsername string) bool {
	return ActiveAt(lease, now) && !req.GetForceNew() && !PlaygroundLeaseNeedsReplacement(req, lease, playgroundAccountID, playgroundUsername)
}

func ReplaceExistingActiveLease(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease, now time.Time, playgroundAccountID string, playgroundUsername string) bool {
	return ActiveAt(lease, now) && !ReuseExistingActiveLease(req, lease, now, playgroundAccountID, playgroundUsername)
}
