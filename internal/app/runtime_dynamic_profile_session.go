package app

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func dynamicProfileConcurrencyHolder(profileID string) string {
	profileID = runtimeSafeID(profileID)
	if profileID == "" {
		profileID = "default"
	}
	return "profile:" + profileID
}

func dynamicProfileSession(profileID string, accountID string, providerID string, endpointID string, input *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySession {
	policy := dynamicProfileSessionPolicy(input, endpointID)
	seed := dynamicProfileSessionSeed(profileID, accountID, providerID, endpointID, policy)
	sessionID := dynamicProfileRequestedSessionID(policy)
	if sessionID == "" {
		sessionID = dynamicProfileSessionID(seed)
	}
	return &proxyruntimev1.ProxySession{
		SessionId:  sessionID,
		ProviderId: strings.TrimSpace(providerID),
		AccountId:  strings.TrimSpace(accountID),
		Purpose:    "in-user-profile",
		Policy:     policy,
	}
}

func dynamicProfileRequestedSessionID(policy *proxyruntimev1.ProxySessionPolicy) string {
	labels := policy.GetLabels()
	return runtimeSafeID(firstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	))
}

func dynamicProfileSessionSeed(profileID string, accountID string, providerID string, endpointID string, policy *proxyruntimev1.ProxySessionPolicy) string {
	return strings.Join([]string{profileID, accountID, providerID, endpointID, dynamicProfilePolicySignature(policy)}, ":")
}

func dynamicProfileSessionPolicy(input *proxyruntimev1.ProxySessionPolicy, endpointID string) *proxyruntimev1.ProxySessionPolicy {
	policy := normalizeDynamicIPSessionPolicy(input)
	policy.Labels["dynamic_ip_endpoint_id"] = strings.TrimSpace(endpointID)
	return policy
}

func dynamicProfilePolicySignature(policy *proxyruntimev1.ProxySessionPolicy) string {
	return strings.Join([]string{
		policy.GetMode().String(),
		policy.GetRotationMode().String(),
		strings.TrimSpace(policy.GetRegion()),
		strings.TrimSpace(policy.GetState()),
		strings.TrimSpace(policy.GetCity()),
		strings.TrimSpace(policy.GetAsn()),
		dynamicIPPolicyDurationText(policy),
	}, "/")
}

func dynamicProfileSessionID(seed string) string {
	return fmt.Sprintf("%08d", hashModulo(seed, 100000000))
}

func dynamicProfileNodeID(profileID string, accountID string, sessionID string, endpointID string, index int) string {
	return runtimeSafeID(fmt.Sprintf("dynamic-%s-%s-%s-%s-%d", profileID, accountID, sessionID, endpointID, index))
}
