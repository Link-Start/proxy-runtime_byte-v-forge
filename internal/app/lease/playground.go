package lease

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func PlaygroundLeaseNeedsReplacement(req *proxygatewayv1.AcquireProxyLeaseRequest, lease *proxygatewayv1.ProxyDynamicLease, accountID string, username string) bool {
	if req.GetAccountId() != strings.TrimSpace(accountID) {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()[LabelProxyUsername]) != strings.TrimSpace(username)
}
