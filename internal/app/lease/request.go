package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func RequestedSessionID(req *proxyruntimev1.AcquireProxyLeaseRequest) string {
	labels := req.GetPolicy().GetLabels()
	return firstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	)
}

func ApplyRequestLabels(req *proxyruntimev1.AcquireProxyLeaseRequest) {
	if req == nil {
		return
	}
	if req.Policy == nil {
		req.Policy = &proxyruntimev1.ProxySessionPolicy{}
	}
	if req.Policy.Labels == nil {
		req.Policy.Labels = map[string]string{}
	}
	req.Policy.Labels["account_id"] = req.GetAccountId()
	req.Policy.Labels["purpose"] = req.GetPurpose()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
