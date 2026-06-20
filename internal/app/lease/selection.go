package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func EndpointID(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	return strings.TrimSpace(appcore.FirstNonEmpty(
		lease.GetSelectionPlan().GetSelectedEndpoint().GetEndpointId(),
		lease.GetEgress().GetLabels()[LabelDynamicIPEndpointID],
		lease.GetSession().GetPolicy().GetLabels()[LabelDynamicIPEndpointID],
	))
}

func SelectedProviderAccountID(plan *proxyruntimev1.ProxyDynamicIPSelectionPlan) string {
	return strings.TrimSpace(plan.GetSelectedEndpoint().GetProviderAccountId())
}

func SelectedDynamicProviderID(plan *proxyruntimev1.ProxyDynamicIPSelectionPlan) string {
	return strings.TrimSpace(plan.GetSelectedEndpoint().GetDynamicProviderId())
}
