package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func dynamicIPPolicyDurationText(policy *proxyruntimev1.ProxySessionPolicy) string {
	if policy == nil || policy.GetStickyTtl() == nil || policy.GetStickyTtl().AsDuration() <= 0 {
		return ""
	}
	return policy.GetStickyTtl().AsDuration().String()
}
