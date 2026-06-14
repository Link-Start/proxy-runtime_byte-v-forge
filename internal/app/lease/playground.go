package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func PlaygroundLeaseNeedsReplacement(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease, accountID string, username string) bool {
	if req.GetAccountId() != strings.TrimSpace(accountID) {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()[LabelProxyUsername]) != strings.TrimSpace(username)
}
