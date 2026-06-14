package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func ConcurrencyPolicy(lease *proxyruntimev1.ProxyDynamicLease) *proxyruntimev1.ProxySessionPolicy {
	if lease == nil {
		return nil
	}
	if policy := lease.GetSession().GetPolicy(); policy != nil {
		return policy
	}
	return &proxyruntimev1.ProxySessionPolicy{RotationMode: lease.GetEgress().GetRotationMode()}
}

func ConcurrencyHolder(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if holder := strings.TrimSpace(lease.GetEgress().GetLabels()["provider_account_concurrency_holder"]); holder != "" {
		return holder
	}
	if holder := strings.TrimSpace(lease.GetSession().GetLabels()["provider_account_concurrency_holder"]); holder != "" {
		return holder
	}
	if leaseID := strings.TrimSpace(lease.GetLeaseId()); leaseID != "" {
		return "lease:" + leaseID
	}
	return ""
}

func DynamicProviderID(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetEgress().GetLabels()["dynamic_provider_id"]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetSession().GetLabels()["dynamic_provider_id"]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if endpoint := lease.GetSelectionPlan().GetSelectedEndpoint(); endpoint != nil {
		return endpoint.GetDynamicProviderId()
	}
	return ""
}
