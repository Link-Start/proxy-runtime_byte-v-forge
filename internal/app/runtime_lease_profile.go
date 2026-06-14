package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func applyLeaseProfileDynamicIPPolicy(settings *runtimeSettingsFile, req *proxyruntimev1.AcquireProxyLeaseRequest) error {
	profile := egressProfileByID(settings, req.GetAccountId())
	if profile == nil {
		if req.GetPurpose() == "in-user-profile" {
			return failedPrecondition("in-user profile dynamic IP is not configured", nil)
		}
		return nil
	}
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return failedPrecondition("in-user profile dynamic IP is not configured", nil)
	}
	if profile.GetExit().GetDynamicIpPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return failedPrecondition("in-user profile lease requires sticky dynamic IP", nil)
	}
	if req.GetPolicy().GetMode() != proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		return invalidArgument("lease request must use sticky dynamic IP", nil)
	}
	req.Policy = profileDynamicIPLeasePolicy(profile.GetExit().GetDynamicIpPolicy(), req.GetPolicy())
	return nil
}

func profileDynamicIPLeasePolicy(profilePolicy *proxyruntimev1.ProxySessionPolicy, requestPolicy *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	policy := normalizeDynamicIPSessionPolicy(profilePolicy)
	request := normalizeDynamicIPSessionPolicy(requestPolicy)
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

func egressProfileByID(settings *runtimeSettingsFile, profileID string) *proxyruntimev1.EgressProfileSettings {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range settings.GetEgressProfiles() {
		if profile.GetProfileId() == profileID {
			return profile
		}
	}
	return nil
}

func playgroundLeaseNeedsReplacement(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease) bool {
	if req.GetAccountId() != playgroundProfileID {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()["proxy_username"]) != playgroundUsername
}
