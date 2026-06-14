package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func EndpointID(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	return strings.TrimSpace(firstNonEmpty(
		lease.GetSelectionPlan().GetSelectedEndpoint().GetEndpointId(),
		lease.GetEgress().GetLabels()["dynamic_ip_endpoint_id"],
		lease.GetSession().GetPolicy().GetLabels()["dynamic_ip_endpoint_id"],
	))
}
