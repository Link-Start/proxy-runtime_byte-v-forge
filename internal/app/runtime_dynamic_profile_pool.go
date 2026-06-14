package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (r *Runtime) dynamicProfilePool(ctx context.Context, settings *runtimeSettingsFile) ([]provider.Node, error) {
	if r.store == nil || r.accountProviders == nil {
		return nil, nil
	}
	settings = normalizeRuntimeSettings(settings)
	instances := dynamicIPProviderInstances(settings)
	if len(instances) == 0 {
		return nil, nil
	}
	accounts, err := r.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	client := r.providerHTTPClient
	endpointHealthScores := r.dynamicIPSelector.dynamicIPEndpointHealthScores(ctx)
	out := []provider.Node{}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() || runtimeSafeID(profile.GetProfileId()) == playgroundProfileID || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		nodes := r.dynamicProfilePoolForProfile(ctx, client, settings, accounts, instances, endpointHealthScores, profile)
		out = append(out, nodes...)
	}
	return out, nil
}

type dynamicProfileEndpointSelection struct {
	account   *proxyruntimev1.ProxyProviderAccount
	accountID string
	config    accountproxy.Config
}

func (r *Runtime) dynamicProfilePoolForProfile(ctx context.Context, client *http.Client, settings *runtimeSettingsFile, accounts []*proxyruntimev1.ProxyProviderAccount, instances []dynamicIPProviderInstance, endpointHealthScores map[string]int, profile *proxyruntimev1.EgressProfileSettings) []provider.Node {
	profileID := runtimeSafeID(profile.GetProfileId())
	exit := profile.GetExit()
	profileDynamicProviderID := runtimeSafeID(exit.GetDynamicProviderId())
	endpointID := dynamicProfileEndpointID(exit)
	policy := dynamicProfileSelectionPolicy(profileID, exit.GetDynamicIpPolicy())
	concurrencyHolder := dynamicProfileConcurrencyHolder(profileID)
	candidates := []scoredDynamicIPEndpointCandidate{}
	selections := map[string]dynamicProfileEndpointSelection{}
	for accountIndex, account := range accounts {
		if account.GetStatus() != proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || !account.GetCredentialConfigured() {
			continue
		}
		if !r.accountProviders.IsSupported(account.GetProviderId()) {
			continue
		}
		accountID := strings.TrimSpace(account.GetAccountId())
		cfg, storedAccountID, err := r.store.ProviderConfig(ctx, accountID)
		if err != nil {
			r.logger.Warn("dynamic profile provider account skipped", "account_id", accountID, "provider_id", account.GetProviderId(), "error", err)
			continue
		}
		accountID = firstNonEmpty(storedAccountID, accountID)
		accountDynamicProviderID := runtimeSafeID(account.GetDynamicProviderId())
		matchedInstances := dynamicProfileProviderInstancesForAccount(instances, account.GetProviderId(), accountDynamicProviderID, profileDynamicProviderID)
		if len(matchedInstances) == 0 {
			continue
		}
		accountCandidates := r.dynamicIPSelector.dynamicIPEndpointCandidatesForAccount(ctx, account, accountIndex, matchedInstances, policy, exit.GetDynamicIpPolicy(), dynamicIPCandidateFilter{concurrencyHolder: concurrencyHolder})
		accountCandidates = dynamicProfileEndpointCandidates(accountCandidates, endpointID)
		for _, candidate := range accountCandidates {
			if candidate.proto == nil {
				continue
			}
			key := dynamicProfileEndpointCandidateKey(candidate)
			candidates = append(candidates, candidate)
			selections[key] = dynamicProfileEndpointSelection{account: account, accountID: accountID, config: cfg}
		}
	}
	applyDynamicIPEndpointHealthScores(candidates, endpointHealthScores)
	selected := chooseDynamicIPEndpointCandidate(candidates, policy, dynamicProfileSelectionKey(profileID, "", endpointID, exit.GetDynamicIpPolicy()), 1)
	selection, ok := selections[dynamicProfileEndpointCandidateKey(selected)]
	if selected.proto == nil || !ok {
		return nil
	}
	limit := dynamicProviderConcurrencyLimit(settings, selected.proto.GetDynamicProviderId(), exit.GetDynamicIpPolicy())
	return r.dynamicProfileNodesForSelection(ctx, client, profile, selection, selected, limit, concurrencyHolder)
}

func (r *Runtime) dynamicProfileNodesForSelection(ctx context.Context, client *http.Client, profile *proxyruntimev1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected scoredDynamicIPEndpointCandidate, concurrencyLimit uint32, concurrencyHolder string) []provider.Node {
	profileID := runtimeSafeID(profile.GetProfileId())
	cfg := selection.config
	cfg.Gateways = []accountproxy.Gateway{selected.endpoint}
	providerClient, err := r.accountProviders.NewSessionProvider(cfg, client)
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error", err)
		return nil
	}
	session := dynamicProfileSession(profileID, selection.accountID, cfg.ProviderID, selected.proto.GetEndpointId(), profile.GetExit().GetDynamicIpPolicy())
	slot, err := r.acquireProviderAccountConcurrencySlot(ctx, selection.account, concurrencyLimit, session.GetPolicy(), concurrencyHolder, leaseConcurrencySlotTTL(session.GetPolicy()))
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error", err)
		return nil
	}
	keepSlot := false
	defer func() {
		if !keepSlot {
			_ = slot.Release(context.Background())
		}
	}()
	nodes, err := providerClient.FetchSession(ctx, session)
	if err != nil {
		r.logger.Warn("dynamic profile provider session skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error", err)
		return nil
	}
	for index, node := range nodes {
		nodes[index] = dynamicProfileLabelNode(node, index, profile, selection, selected)
	}
	keepSlot = true
	return nodes
}

func dynamicProfileConcurrencyHolder(profileID string) string {
	profileID = runtimeSafeID(profileID)
	if profileID == "" {
		profileID = "default"
	}
	return "profile:" + profileID
}

func dynamicProfileLabelNode(node provider.Node, index int, profile *proxyruntimev1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected scoredDynamicIPEndpointCandidate) provider.Node {
	profileID := runtimeSafeID(profile.GetProfileId())
	policy := dynamicProfileSessionPolicy(profile.GetExit().GetDynamicIpPolicy(), selected.proto.GetEndpointId())
	node.ID = dynamicProfileNodeID(profileID, selection.accountID, node.SessionID, selected.proto.GetEndpointId(), index)
	node.ProviderID = selection.config.ProviderID
	node.Labels = cloneLabels(node.Labels)
	node.Labels["egress_profile_id"] = profileID
	node.Labels["egress_profile_display_name"] = strings.TrimSpace(profile.GetDisplayName())
	node.Labels["line_kind"] = egressProfileLineKind(profile.GetLine().GetKind())
	node.Labels["line_resource_id"] = strings.TrimSpace(profile.GetLine().GetMihomoNode().GetResourceId())
	node.Labels["line_node_id"] = strings.TrimSpace(profile.GetLine().GetMihomoNode().GetNodeId())
	node.Labels["provider_account_id"] = selection.accountID
	node.Labels["provider_account_display_name"] = strings.TrimSpace(selection.account.GetDisplayName())
	node.Labels["provider_id"] = selection.config.ProviderID
	node.Labels["session_id"] = strings.TrimSpace(node.SessionID)
	node.Labels["dynamic_provider_id"] = selected.proto.GetDynamicProviderId()
	node.Labels["dynamic_provider_ids"] = selected.proto.GetDynamicProviderId()
	node.Labels["dynamic_ip_endpoint_id"] = selected.proto.GetEndpointId()
	node.Labels["dynamic_ip_endpoint_url"] = selected.proto.GetEndpointUrl()
	node.Labels["session_mode"] = policy.GetMode().String()
	node.Labels["rotation_mode"] = policy.GetRotationMode().String()
	node.Labels["region"] = strings.TrimSpace(policy.GetRegion())
	node.Labels["state"] = strings.TrimSpace(policy.GetState())
	node.Labels["city"] = strings.TrimSpace(policy.GetCity())
	node.Labels["asn"] = strings.TrimSpace(policy.GetAsn())
	node.Labels["sticky_ttl"] = dynamicIPPolicyDurationText(policy)
	node.Labels["mode"] = "dynamic_profile"
	return node
}

func dynamicProfileEndpointCandidateKey(candidate scoredDynamicIPEndpointCandidate) string {
	if candidate.proto == nil {
		return ""
	}
	return strings.Join([]string{
		candidate.proto.GetProviderAccountId(),
		candidate.proto.GetProviderId(),
		candidate.proto.GetDynamicProviderId(),
		candidate.proto.GetEndpointId(),
		candidate.proto.GetEndpointUrl(),
	}, "/")
}

func dynamicProfileProviderInstancesForAccount(instances []dynamicIPProviderInstance, providerID string, accountDynamicProviderID string, profileDynamicProviderID string) []dynamicIPProviderInstance {
	providerID = strings.TrimSpace(providerID)
	accountDynamicProviderID = runtimeSafeID(accountDynamicProviderID)
	profileDynamicProviderID = runtimeSafeID(profileDynamicProviderID)
	out := []dynamicIPProviderInstance{}
	for _, instance := range instances {
		if strings.TrimSpace(instance.providerID) != providerID {
			continue
		}
		if accountDynamicProviderID != "" && accountDynamicProviderID != instance.dynamicProviderID {
			continue
		}
		if profileDynamicProviderID != "" && profileDynamicProviderID != instance.dynamicProviderID {
			continue
		}
		out = append(out, instance)
	}
	return out
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
