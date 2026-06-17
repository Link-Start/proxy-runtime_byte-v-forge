package lease

import (
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var (
	ErrProfileDynamicIPNotConfigured  = errors.New("in-user profile dynamic IP is not configured")
	ErrProfileLeaseRequiresSticky     = errors.New("in-user profile lease requires sticky dynamic IP")
	ErrRequestRequiresStickyDynamicIP = errors.New("lease request must use sticky dynamic IP")
)

func ApplyProfileDynamicIPPolicy(profiles []*proxyruntimev1.EgressProfileSettings, req *proxyruntimev1.AcquireProxyLeaseRequest) error {
	profile := EgressProfileByID(profiles, req.GetAccountId())
	if profile == nil {
		if req.GetPurpose() == "in-user-profile" {
			return ErrProfileDynamicIPNotConfigured
		}
		return nil
	}
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return ErrProfileDynamicIPNotConfigured
	}
	if profile.GetExit().GetDynamicIpPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return ErrProfileLeaseRequiresSticky
	}
	if req.GetPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return ErrRequestRequiresStickyDynamicIP
	}
	req.Policy = ProfileDynamicIPLeasePolicy(profile.GetExit().GetDynamicIpPolicy(), req.GetPolicy())
	return nil
}

func ResolveAcquireRequestAccountID(profiles []*proxyruntimev1.EgressProfileSettings, rules []*proxyruntimev1.ProxyIngressRuleSettings, req *proxyruntimev1.AcquireProxyLeaseRequest) {
	if req == nil {
		return
	}
	accountID := strings.TrimSpace(req.GetAccountId())
	if accountID == "" || EgressProfileByID(profiles, accountID) != nil {
		return
	}
	rule := IngressRuleByUsername(rules, accountID)
	if rule == nil || strings.TrimSpace(rule.GetProfileId()) == "" {
		return
	}
	req.AccountId = strings.TrimSpace(rule.GetProfileId())
}

func ProfileDynamicIPLeasePolicy(profilePolicy *proxyruntimev1.ProxySessionPolicy, requestPolicy *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	policy := NormalizeDynamicIPSessionPolicy(profilePolicy)
	request := NormalizeDynamicIPSessionPolicy(requestPolicy)
	policy.StickyTtl = cloneDuration(request.GetStickyTtl())
	policy.Labels = cloneStringMap(policy.GetLabels())
	if policy.Labels == nil {
		policy.Labels = map[string]string{}
	}
	for key, value := range request.GetLabels() {
		policy.Labels[key] = value
	}
	return policy
}

func EgressProfileByID(profiles []*proxyruntimev1.EgressProfileSettings, profileID string) *proxyruntimev1.EgressProfileSettings {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range profiles {
		if profile.GetProfileId() == profileID {
			return profile
		}
	}
	return nil
}
