package lease

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

const (
	LabelAccountID                        = "account_id"
	LabelPurpose                          = "purpose"
	LabelSessionID                        = "session_id"
	LabelProviderAccountID                = "provider_account_id"
	LabelProviderAccountConcurrencyHolder = "provider_account_concurrency_holder"
	LabelDynamicProviderID                = "dynamic_provider_id"
	LabelDynamicIPEndpointID              = "dynamic_ip_endpoint_id"
	LabelSelectionID                      = "selection_id"
)

func ConcurrencyMode(policy *proxyruntimev1.ProxySessionPolicy) proxyruntimev1.ProxySessionMode {
	if policy == nil {
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
	if policy.GetMode() == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING || policy.GetRotationMode() == proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST {
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	}
	return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
}

func ConcurrencyModeText(policy *proxyruntimev1.ProxySessionPolicy) string {
	if ConcurrencyMode(policy) == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return "rotating"
	}
	return "sticky"
}

func ConcurrencySlotTTL(policy *proxyruntimev1.ProxySessionPolicy, defaultTTL time.Duration, buffer time.Duration) time.Duration {
	ttl := defaultTTL
	if policy != nil && policy.GetStickyTtl() != nil && policy.GetStickyTtl().AsDuration() > 0 {
		ttl = policy.GetStickyTtl().AsDuration()
	}
	return ttl + buffer
}

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
	if holder := strings.TrimSpace(lease.GetEgress().GetLabels()[LabelProviderAccountConcurrencyHolder]); holder != "" {
		return holder
	}
	if holder := strings.TrimSpace(lease.GetSession().GetLabels()[LabelProviderAccountConcurrencyHolder]); holder != "" {
		return holder
	}
	return HolderForLeaseID(lease.GetLeaseId())
}

func HolderForLeaseID(leaseID string) string {
	if leaseID = strings.TrimSpace(leaseID); leaseID != "" {
		return "lease:" + leaseID
	}
	return ""
}

func DynamicProviderID(lease *proxyruntimev1.ProxyDynamicLease) string {
	if lease == nil {
		return ""
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetEgress().GetLabels()[LabelDynamicProviderID]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if dynamicProviderID := strings.TrimSpace(lease.GetSession().GetLabels()[LabelDynamicProviderID]); dynamicProviderID != "" {
		return dynamicProviderID
	}
	if endpoint := lease.GetSelectionPlan().GetSelectedEndpoint(); endpoint != nil {
		return endpoint.GetDynamicProviderId()
	}
	return ""
}
