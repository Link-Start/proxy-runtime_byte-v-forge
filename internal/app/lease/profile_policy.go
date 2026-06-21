package lease

import (
	"errors"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

var (
	ErrProfileDynamicIPNotConfigured  = errors.New("in-user profile dynamic IP is not configured")
	ErrProfileLeaseRequiresSticky     = errors.New("in-user profile lease requires sticky dynamic IP")
	ErrRequestRequiresStickyDynamicIP = errors.New("lease request must use sticky dynamic IP")
)

func ApplyProfileDynamicIPPolicy(profiles []*proxygatewayv1.EgressProfileSettings, req *proxygatewayv1.AcquireProxyLeaseRequest) error {
	profile := EgressProfileByID(profiles, req.GetAccountId())
	if profile == nil {
		if req.GetPurpose() == "in-user-profile" {
			return ErrProfileDynamicIPNotConfigured
		}
		return nil
	}
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return ErrProfileDynamicIPNotConfigured
	}
	if profile.GetExit().GetDynamicIpPolicy().GetMode() != proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return ErrProfileLeaseRequiresSticky
	}
	if req.GetPolicy().GetMode() != proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return ErrRequestRequiresStickyDynamicIP
	}
	req.Policy = ProfileDynamicIPLeasePolicy(profile.GetExit().GetDynamicIpPolicy(), req.GetPolicy())
	return nil
}

func ResolveAcquireRequestAccountID(profiles []*proxygatewayv1.EgressProfileSettings, rules []*proxygatewayv1.ProxyIngressRuleSettings, req *proxygatewayv1.AcquireProxyLeaseRequest) {
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

func ProfileDynamicIPLeasePolicy(profilePolicy *proxygatewayv1.ProxySessionPolicy, requestPolicy *proxygatewayv1.ProxySessionPolicy) *proxygatewayv1.ProxySessionPolicy {
	policy := kernel.NormalizeDynamicIPSessionPolicy(profilePolicy)
	request := kernel.NormalizeDynamicIPSessionPolicy(requestPolicy)
	policy.StickyTtl = appcore.CloneDuration(request.GetStickyTtl())
	policy.Labels = appcore.CloneStringMap(policy.GetLabels())
	if policy.Labels == nil {
		policy.Labels = map[string]string{}
	}
	for key, value := range request.GetLabels() {
		policy.Labels[key] = value
	}
	return policy
}

func EgressProfileByID(profiles []*proxygatewayv1.EgressProfileSettings, profileID string) *proxygatewayv1.EgressProfileSettings {
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
