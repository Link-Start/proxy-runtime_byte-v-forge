package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func dynamicProfileEndpointID(exit *proxyruntimev1.EgressProfileExitSettings) string {
	return strings.TrimSpace(exit.GetDynamicIpPolicy().GetLabels()["dynamic_ip_endpoint_id"])
}

func dynamicProfileEndpointCandidates(candidates []scoredDynamicIPEndpointCandidate, endpointID string) []scoredDynamicIPEndpointCandidate {
	endpointID = strings.TrimSpace(endpointID)
	if endpointID == "" {
		return candidates
	}
	out := []scoredDynamicIPEndpointCandidate{}
	for _, candidate := range candidates {
		if candidate.proto.GetEndpointId() == endpointID {
			out = append(out, candidate)
		}
	}
	return out
}

func dynamicProfileSelectionPolicy(profileID string, policy *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxyDynamicIPSelectionPolicy {
	return leaseapp.NormalizeDynamicIPSelectionPolicy(&proxyruntimev1.AcquireProxyLeaseRequest{
		AccountId: strings.TrimSpace(profileID),
		Purpose:   "in-user-profile",
		Policy:    policy,
	})
}

func dynamicProfileSelectionKey(profileID string, accountID string, endpointID string, policy *proxyruntimev1.ProxySessionPolicy) string {
	labels := policy.GetLabels()
	return strings.Join([]string{
		appcore.FirstNonEmpty(labels["selection_seed"], labels["proxy_selection_seed"], profileID),
		strings.TrimSpace(accountID),
		strings.TrimSpace(endpointID),
		dynamicProfilePolicySignature(dynamicProfileSessionPolicy(policy, endpointID)),
	}, ":")
}
