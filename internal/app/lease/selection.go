package lease

import (
	"strconv"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/geox"
)

func EndpointID(lease *proxygatewayv1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	return strings.TrimSpace(appcore.FirstNonEmpty(
		lease.GetSelectionPlan().GetSelectedEndpoint().GetEndpointId(),
		lease.GetEgress().GetLabels()[LabelDynamicIPEndpointID],
		lease.GetSession().GetPolicy().GetLabels()[LabelDynamicIPEndpointID],
	))
}

func SelectedProviderAccountID(plan *proxygatewayv1.ProxyDynamicIPSelectionPlan) string {
	return strings.TrimSpace(plan.GetSelectedEndpoint().GetProviderAccountId())
}

func SelectedDynamicProviderID(plan *proxygatewayv1.ProxyDynamicIPSelectionPlan) string {
	return strings.TrimSpace(plan.GetSelectedEndpoint().GetDynamicProviderId())
}

const maxDynamicIPSelectionAttempts = 20

func NormalizeDynamicIPSelectionPolicy(req *proxygatewayv1.AcquireProxyLeaseRequest) *proxygatewayv1.ProxyDynamicIPSelectionPolicy {
	in := req.GetSelectionPolicy()
	policy := &proxygatewayv1.ProxyDynamicIPSelectionPolicy{}
	if in != nil {
		policy.CountryCode = strings.TrimSpace(in.GetCountryCode())
		policy.Region = strings.TrimSpace(in.GetRegion())
		policy.Purpose = strings.TrimSpace(in.GetPurpose())
		policy.MaxAttempts = in.GetMaxAttempts()
	}
	if policy.CountryCode == "" {
		policy.CountryCode = appcore.FirstNonEmpty(req.GetPolicy().GetLabels()["country_code"], req.GetPolicy().GetRegion())
	}
	if policy.Region == "" {
		policy.Region = appcore.FirstNonEmpty(req.GetPolicy().GetLabels()["region"], req.GetPolicy().GetRegion())
	}
	if policy.Purpose == "" {
		policy.Purpose = strings.TrimSpace(req.GetPurpose())
	}
	policy.CountryCode = geox.NormalizeCountryAlpha2(policy.CountryCode)
	policy.Region = strings.ToUpper(strings.TrimSpace(policy.Region))
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = 10
	}
	if policy.MaxAttempts > maxDynamicIPSelectionAttempts {
		policy.MaxAttempts = maxDynamicIPSelectionAttempts
	}
	return policy
}

func DynamicIPSelectionAttempt(req *proxygatewayv1.AcquireProxyLeaseRequest) int {
	if req == nil || req.GetPolicy() == nil {
		return 1
	}
	value := strings.TrimSpace(req.GetPolicy().GetLabels()[LabelAttempt])
	if value == "" {
		return 1
	}
	attempt, err := strconv.Atoi(value)
	if err != nil || attempt < 1 {
		return 1
	}
	return attempt
}

func DynamicIPSelectionMaxAttempts(policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) int {
	attempts := int(policy.GetMaxAttempts())
	if attempts < 1 {
		return 1
	}
	if attempts > maxDynamicIPSelectionAttempts {
		return maxDynamicIPSelectionAttempts
	}
	return attempts
}

func DynamicIPSelectionKey(req *proxygatewayv1.AcquireProxyLeaseRequest) string {
	if req == nil {
		return ""
	}
	labels := req.GetPolicy().GetLabels()
	return appcore.FirstNonEmpty(
		labels["selection_seed"],
		labels["proxy_selection_seed"],
		labels["job_id"],
		req.GetAccountId(),
		req.GetPurpose(),
	)
}
