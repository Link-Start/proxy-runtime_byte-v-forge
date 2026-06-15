package app

import (
	"context"
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
			r.logger.Warn("dynamic profile provider account skipped", "account_id", accountID, "provider_id", account.GetProviderId(), "error_type", errorLogType(err))
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
