package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	dashboardapp "github.com/byte-v-forge/proxy-gateway/internal/app/dashboard"
	"github.com/byte-v-forge/proxy-gateway/internal/app/dynamic"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-gateway/internal/runtimehttp"
)

func (r *Runtime) refreshDynamicProfileSelectionMetadata(ctx context.Context) {
	settings, err := r.settings.Load(ctx)
	if err != nil {
		return
	}
	nodes, _ := r.dynamicProfilePoolSnapshot()
	if len(nodes) == 0 {
		return
	}
	for _, profile := range settings.GetEgressProfiles() {
		r.applyDynamicProfileSelectionMetadata(ctx, nodes, profile)
	}
	r.setDynamicProfilePoolSnapshot(nodes)
}

func (r *Runtime) applyDynamicProfileSelectionMetadata(ctx context.Context, nodes []provider.Node, profile *proxygatewayv1.EgressProfileSettings) {
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return
	}
	clearDynamicProfileExitIP(nodes, profile.GetProfileId())
	groupName := strings.TrimSpace(profile.GetDisplayName())
	selected, err := r.mihomoProxyGroupSelected(ctx, groupName)
	if err != nil || selected == "" {
		return
	}
	markDynamicProfileSelection(nodes, profile.GetProfileId(), selected)
}

func clearDynamicProfileExitIP(nodes []provider.Node, profileID string) {
	profileID = appcore.RuntimeSafeID(profileID)
	for index := range nodes {
		if strings.TrimSpace(nodes[index].Labels["egress_profile_id"]) != profileID {
			continue
		}
		nodes[index].Labels = cloneLabels(nodes[index].Labels)
		delete(nodes[index].Labels, "exit_ip")
		delete(nodes[index].Labels, "selected")
		delete(nodes[index].Labels, "mihomo_proxy_name")
	}
}

func markDynamicProfileSelection(nodes []provider.Node, profileID string, selectedProxy string) {
	index := dynamicProfileNodeIndex(nodes, profileID, selectedProxy)
	if index < 0 {
		return
	}
	nodes[index].Labels = cloneLabels(nodes[index].Labels)
	nodes[index].Labels["selected"] = "true"
	nodes[index].Labels["mihomo_proxy_name"] = strings.TrimSpace(selectedProxy)
}

func dynamicProfileNodeIndex(nodes []provider.Node, profileID string, selectedProxy string) int {
	profileID = appcore.RuntimeSafeID(profileID)
	for index, node := range nodes {
		if strings.TrimSpace(node.Labels["egress_profile_id"]) != profileID {
			continue
		}
		nodeID := strings.TrimSpace(node.ID)
		if nodeID != "" && strings.Contains(selectedProxy, nodeID) {
			return index
		}
	}
	return -1
}

func (r *Runtime) mihomoProxyGroupNow(ctx context.Context, groupName string) (string, error) {
	group, err := r.mihomoProxyGroup(ctx, groupName)
	return group.Now, err
}

func (r *Runtime) mihomoProxyGroupSelected(ctx context.Context, groupName string) (string, error) {
	group, err := r.mihomoProxyGroup(ctx, groupName)
	if err != nil {
		return "", err
	}
	if group.Now != "" {
		return group.Now, nil
	}
	if len(group.All) == 1 {
		return group.All[0], nil
	}
	return "", nil
}

type mihomoProxyGroupState struct {
	Now string   `json:"now"`
	All []string `json:"all"`
}

func (r *Runtime) mihomoProxyGroup(ctx context.Context, groupName string) (mihomoProxyGroupState, error) {
	target, err := dashboardapp.APIURL(r.cfg.Mihomo.APIAddr)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	target.Path = "/proxies/" + url.PathEscape(groupName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	dashboardapp.ApplyControllerAuthorization(req, r.cfg.ControlAuthToken)
	resp, err := runtimehttp.New(r.cfg.RequestTimeout).Do(req)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	defer resp.Body.Close()
	var payload mihomoProxyGroupState
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return mihomoProxyGroupState{}, err
	}
	payload.Now = strings.TrimSpace(payload.Now)
	payload.All = compactStrings(payload.All)
	return payload, nil
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

const dynamicProfileSlotReleaseTimeout = 5 * time.Second

func (r *Runtime) dynamicProfileNodesForSelection(ctx context.Context, client *http.Client, profile *proxygatewayv1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected dynamic.ScoredEndpointCandidate, concurrencyLimit uint32, concurrencyHolder string) []provider.Node {
	profileID := appcore.RuntimeSafeID(profile.GetProfileId())
	cfg := selection.config
	cfg.Gateways = []accountproxy.Gateway{selected.Endpoint}
	providerClient, err := r.accountProviders.NewSessionProvider(cfg, client, r.clock)
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	session := dynamicProfileSession(profileID, selection.accountID, cfg.ProviderID, selected.Proto.GetEndpointId(), profile.GetExit().GetDynamicIpPolicy())
	slot, err := leaseapp.AcquireProviderAccountConcurrencySlot(ctx, r.providerConcurrency, selection.account.GetAccountId(), concurrencyLimit, session.GetPolicy(), concurrencyHolder, leaseapp.ConcurrencySlotTTL(session.GetPolicy(), kernel.DefaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	keepSlot := false
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dynamicProfileSlotReleaseTimeout)
		defer cancel()
		_ = leaseapp.ReleaseConcurrencySlotUnlessKept(releaseCtx, slot, keepSlot)
	}()
	nodes, err := leaseapp.FetchProviderSession(ctx, providerClient, session)
	if err != nil {
		r.logger.Warn("dynamic profile provider session skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	for index, node := range nodes {
		nodes[index] = dynamicProfileLabelNode(node, index, profile, selection, selected)
	}
	keepSlot = true
	return nodes
}

func dynamicProfileLabelNode(node provider.Node, index int, profile *proxygatewayv1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected dynamic.ScoredEndpointCandidate) provider.Node {
	profileID := appcore.RuntimeSafeID(profile.GetProfileId())
	policy := dynamicProfileSessionPolicy(profile.GetExit().GetDynamicIpPolicy(), selected.Proto.GetEndpointId())
	node.ID = dynamicProfileNodeID(profileID, selection.accountID, node.SessionID, selected.Proto.GetEndpointId(), index)
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
	node.Labels["dynamic_provider_id"] = selected.Proto.GetDynamicProviderId()
	node.Labels["dynamic_provider_ids"] = selected.Proto.GetDynamicProviderId()
	node.Labels["dynamic_ip_endpoint_id"] = selected.Proto.GetEndpointId()
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

func (r *Runtime) dynamicProfilePool(ctx context.Context, settings *runtimeSettingsFile) ([]provider.Node, error) {
	if r.store == nil || r.accountProviders == nil {
		return nil, nil
	}
	settings = kernel.NormalizeRuntimeSettings(settings)
	instances := dynamic.ProviderInstances(settings)
	if len(instances) == 0 {
		return nil, nil
	}
	accounts, err := r.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	client := r.providerHTTPClient
	endpointHealthScores := r.dynamicIPSelector.DynamicIPEndpointHealthScores(ctx)
	out := []provider.Node{}
	for _, profile := range settings.GetEgressProfiles() {
		if !profile.GetEnabled() || appcore.RuntimeSafeID(profile.GetProfileId()) == kernel.PlaygroundProfileID || profile.GetExit().GetKind() != proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
			continue
		}
		nodes := r.dynamicProfilePoolForProfile(ctx, client, settings, accounts, instances, endpointHealthScores, profile)
		out = append(out, nodes...)
	}
	return out, nil
}

type dynamicProfileEndpointSelection struct {
	account   *proxygatewayv1.ProxyProviderAccount
	accountID string
	config    accountproxy.Config
}

func (r *Runtime) dynamicProfilePoolForProfile(ctx context.Context, client *http.Client, settings *runtimeSettingsFile, accounts []*proxygatewayv1.ProxyProviderAccount, instances []dynamic.ProviderInstance, endpointHealthScores map[string]int, profile *proxygatewayv1.EgressProfileSettings) []provider.Node {
	profileID := appcore.RuntimeSafeID(profile.GetProfileId())
	exit := profile.GetExit()
	profileDynamicProviderID := appcore.RuntimeSafeID(exit.GetDynamicProviderId())
	endpointID := dynamicProfileEndpointID(exit)
	policy := dynamicProfileSelectionPolicy(profileID, exit.GetDynamicIpPolicy())
	concurrencyHolder := dynamicProfileConcurrencyHolder(profileID)
	candidates := []dynamic.ScoredEndpointCandidate{}
	selections := map[string]dynamicProfileEndpointSelection{}
	for accountIndex, account := range accounts {
		if account.GetStatus() != proxygatewayv1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || !account.GetCredentialConfigured() {
			continue
		}
		if !r.accountProviders.IsSupported(account.GetProviderId()) {
			continue
		}
		accountID := strings.TrimSpace(account.GetAccountId())
		cfg, storedAccountID, err := r.store.ProviderConfig(ctx, accountID)
		if err != nil {
			r.logger.Warn("dynamic profile provider account skipped", "account_id", accountID, "provider_id", account.GetProviderId(), "error_type", appcore.ErrorLogType(err))
			continue
		}
		accountID = appcore.FirstNonEmpty(storedAccountID, accountID)
		accountDynamicProviderID := appcore.RuntimeSafeID(account.GetDynamicProviderId())
		matchedInstances := dynamicProfileProviderInstancesForAccount(instances, account.GetProviderId(), accountDynamicProviderID, profileDynamicProviderID)
		if len(matchedInstances) == 0 {
			continue
		}
		accountCandidates := r.dynamicIPSelector.DynamicIPEndpointCandidatesForAccount(ctx, account, accountIndex, matchedInstances, policy, exit.GetDynamicIpPolicy(), dynamic.IPCandidateFilter{ConcurrencyHolder: concurrencyHolder})
		accountCandidates = dynamicProfileEndpointCandidates(accountCandidates, endpointID)
		for _, candidate := range accountCandidates {
			if candidate.Proto == nil {
				continue
			}
			key := dynamicProfileEndpointCandidateKey(candidate)
			candidates = append(candidates, candidate)
			selections[key] = dynamicProfileEndpointSelection{account: account, accountID: accountID, config: cfg}
		}
	}
	dynamic.ApplyEndpointHealthScores(candidates, endpointHealthScores)
	selected := dynamic.ChooseEndpointCandidate(candidates, policy, dynamicProfileSelectionKey(profileID, "", endpointID, exit.GetDynamicIpPolicy()), 1)
	selection, ok := selections[dynamicProfileEndpointCandidateKey(selected)]
	if selected.Proto == nil || !ok {
		return nil
	}
	limit := dynamicProviderConcurrencyLimit(settings, selected.Proto.GetDynamicProviderId(), exit.GetDynamicIpPolicy())
	return r.dynamicProfileNodesForSelection(ctx, client, profile, selection, selected, limit, concurrencyHolder)
}

func dynamicProfileEndpointCandidateKey(candidate dynamic.ScoredEndpointCandidate) string {
	if candidate.Proto == nil {
		return ""
	}
	return strings.Join([]string{
		candidate.Proto.GetProviderAccountId(),
		candidate.Proto.GetProviderId(),
		candidate.Proto.GetDynamicProviderId(),
		candidate.Proto.GetEndpointId(),
		candidate.Proto.GetEndpointUrl(),
	}, "/")
}

func dynamicProfileProviderInstancesForAccount(instances []dynamic.ProviderInstance, providerID string, accountDynamicProviderID string, profileDynamicProviderID string) []dynamic.ProviderInstance {
	providerID = strings.TrimSpace(providerID)
	accountDynamicProviderID = appcore.RuntimeSafeID(accountDynamicProviderID)
	profileDynamicProviderID = appcore.RuntimeSafeID(profileDynamicProviderID)
	out := []dynamic.ProviderInstance{}
	for _, instance := range instances {
		if strings.TrimSpace(instance.ProviderID) != providerID {
			continue
		}
		if accountDynamicProviderID != "" && accountDynamicProviderID != instance.DynamicProviderID {
			continue
		}
		if profileDynamicProviderID != "" && profileDynamicProviderID != instance.DynamicProviderID {
			continue
		}
		out = append(out, instance)
	}
	return out
}

func dynamicProfileEndpointID(exit *proxygatewayv1.EgressProfileExitSettings) string {
	return strings.TrimSpace(exit.GetDynamicIpPolicy().GetLabels()["dynamic_ip_endpoint_id"])
}

func dynamicProfileEndpointCandidates(candidates []dynamic.ScoredEndpointCandidate, endpointID string) []dynamic.ScoredEndpointCandidate {
	endpointID = strings.TrimSpace(endpointID)
	if endpointID == "" {
		return candidates
	}
	out := []dynamic.ScoredEndpointCandidate{}
	for _, candidate := range candidates {
		if candidate.Proto.GetEndpointId() == endpointID {
			out = append(out, candidate)
		}
	}
	return out
}

func dynamicProfileSelectionPolicy(profileID string, policy *proxygatewayv1.ProxySessionPolicy) *proxygatewayv1.ProxyDynamicIPSelectionPolicy {
	return leaseapp.NormalizeDynamicIPSelectionPolicy(&proxygatewayv1.AcquireProxyLeaseRequest{
		AccountId: strings.TrimSpace(profileID),
		Purpose:   "in-user-profile",
		Policy:    policy,
	})
}

func dynamicProfileSelectionKey(profileID string, accountID string, endpointID string, policy *proxygatewayv1.ProxySessionPolicy) string {
	labels := policy.GetLabels()
	return strings.Join([]string{
		appcore.FirstNonEmpty(labels["selection_seed"], labels["proxy_selection_seed"], profileID),
		strings.TrimSpace(accountID),
		strings.TrimSpace(endpointID),
		dynamicProfilePolicySignature(dynamicProfileSessionPolicy(policy, endpointID)),
	}, ":")
}

func dynamicProfileConcurrencyHolder(profileID string) string {
	profileID = appcore.RuntimeSafeID(profileID)
	if profileID == "" {
		profileID = "default"
	}
	return "profile:" + profileID
}

func dynamicProfileSession(profileID string, accountID string, providerID string, endpointID string, input *proxygatewayv1.ProxySessionPolicy) *proxygatewayv1.ProxySession {
	policy := dynamicProfileSessionPolicy(input, endpointID)
	seed := dynamicProfileSessionSeed(profileID, accountID, providerID, endpointID, policy)
	sessionID := dynamicProfileRequestedSessionID(policy)
	if sessionID == "" {
		sessionID = dynamicProfileSessionID(seed)
	}
	return &proxygatewayv1.ProxySession{
		SessionId:  sessionID,
		ProviderId: strings.TrimSpace(providerID),
		AccountId:  strings.TrimSpace(accountID),
		Purpose:    "in-user-profile",
		Policy:     policy,
	}
}

func dynamicProfileRequestedSessionID(policy *proxygatewayv1.ProxySessionPolicy) string {
	labels := policy.GetLabels()
	return appcore.RuntimeSafeID(appcore.FirstNonEmpty(
		labels["session_id"],
		labels["sticky_session_id"],
		labels["sticky_id"],
		labels["sid"],
		labels["session"],
	))
}

func dynamicProfileSessionSeed(profileID string, accountID string, providerID string, endpointID string, policy *proxygatewayv1.ProxySessionPolicy) string {
	return strings.Join([]string{profileID, accountID, providerID, endpointID, dynamicProfilePolicySignature(policy)}, ":")
}

func dynamicProfileSessionPolicy(input *proxygatewayv1.ProxySessionPolicy, endpointID string) *proxygatewayv1.ProxySessionPolicy {
	policy := kernel.NormalizeDynamicIPSessionPolicy(input)
	policy.Labels["dynamic_ip_endpoint_id"] = strings.TrimSpace(endpointID)
	return policy
}

func dynamicProfilePolicySignature(policy *proxygatewayv1.ProxySessionPolicy) string {
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
	return fmt.Sprintf("%08d", appcore.HashModulo(seed, 100000000))
}

func dynamicProfileNodeID(profileID string, accountID string, sessionID string, endpointID string, index int) string {
	return appcore.RuntimeSafeID(fmt.Sprintf("dynamic-%s-%s-%s-%s-%d", profileID, accountID, sessionID, endpointID, index))
}

func (r *Runtime) setDynamicProfilePoolSnapshot(nodes []provider.Node) {
	r.dynamicProfileMu.Lock()
	defer r.dynamicProfileMu.Unlock()
	r.dynamicProfilePoolNodes = cloneProviderNodes(nodes)
	r.dynamicProfileUpdatedAt = r.clock.Now().UTC()
}

func (r *Runtime) dynamicProfilePoolSnapshot() ([]provider.Node, time.Time) {
	r.dynamicProfileMu.RLock()
	defer r.dynamicProfileMu.RUnlock()
	return cloneProviderNodes(r.dynamicProfilePoolNodes), r.dynamicProfileUpdatedAt
}

func cloneProviderNodes(nodes []provider.Node) []provider.Node {
	out := make([]provider.Node, 0, len(nodes))
	for _, node := range nodes {
		item := node
		if node.URL != nil {
			copied := *node.URL
			item.URL = &copied
		}
		item.Labels = cloneLabels(node.Labels)
		out = append(out, item)
	}
	return out
}
