package app

import (
	"context"
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/dynamic"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func (r *Runtime) dynamicProfilePool(ctx context.Context, settings *runtimeSettingsFile) ([]provider.Node, error) {
	if r.store == nil || r.accountProviders == nil {
		return nil, nil
	}
	settings = settingscore.NormalizeRuntimeSettings(settings)
	instances := dynamic.ProviderInstances(settings)
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
		if !profile.GetEnabled() || appcore.RuntimeSafeID(profile.GetProfileId()) == playgroundProfileID || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
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

func (r *Runtime) dynamicProfilePoolForProfile(ctx context.Context, client *http.Client, settings *runtimeSettingsFile, accounts []*proxyruntimev1.ProxyProviderAccount, instances []dynamic.ProviderInstance, endpointHealthScores map[string]int, profile *proxyruntimev1.EgressProfileSettings) []provider.Node {
	profileID := appcore.RuntimeSafeID(profile.GetProfileId())
	exit := profile.GetExit()
	profileDynamicProviderID := appcore.RuntimeSafeID(exit.GetDynamicProviderId())
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
			r.logger.Warn("dynamic profile provider account skipped", "account_id", accountID, "provider_id", account.GetProviderId(), "error_type", appcore.ErrorLogType(err))
			continue
		}
		accountID = appcore.FirstNonEmpty(storedAccountID, accountID)
		accountDynamicProviderID := appcore.RuntimeSafeID(account.GetDynamicProviderId())
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
