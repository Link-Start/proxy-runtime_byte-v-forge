package app

import (
	"context"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (p *dynamicIPSelector) dynamicIPEndpointCandidates(ctx context.Context, settings *runtimeSettingsFile, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, sessionPolicy *proxyruntimev1.ProxySessionPolicy) ([]scoredDynamicIPEndpointCandidate, error) {
	if p == nil || p.store == nil {
		return nil, internalError("dynamic IP selection store is not configured", nil)
	}
	accounts, err := p.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	providerInstances := dynamicIPProviderInstances(settings)
	filter := dynamicIPCandidateFilterFromPolicy(sessionPolicy)
	out := make([]scoredDynamicIPEndpointCandidate, 0)
	for accountIndex, account := range accounts {
		if account.GetStatus() != proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || !account.GetCredentialConfigured() {
			continue
		}
		if !p.providerSupported(account.GetProviderId()) {
			continue
		}
		out = append(out, p.dynamicIPEndpointCandidatesForAccount(ctx, account, accountIndex, providerInstances, policy, sessionPolicy, filter)...)
	}
	applyDynamicIPEndpointHealthScores(out, p.dynamicIPEndpointHealthScores(ctx))
	return out, nil
}

type dynamicIPCandidateFilter struct {
	dynamicProviderID string
	endpointID        string
	concurrencyHolder string
}

func dynamicIPCandidateFilterFromPolicy(policy *proxyruntimev1.ProxySessionPolicy) dynamicIPCandidateFilter {
	labels := policy.GetLabels()
	return dynamicIPCandidateFilter{
		dynamicProviderID: runtimeSafeID(labels["dynamic_provider_id"]),
		endpointID:        strings.TrimSpace(labels["dynamic_ip_endpoint_id"]),
	}
}

func (p *dynamicIPSelector) providerAccountConcurrencyAvailable(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, provider dynamicIPProviderInstance, policy *proxyruntimev1.ProxySessionPolicy, holder string) (bool, error) {
	if p.concurrency == nil {
		return true, nil
	}
	return p.concurrency.Available(ctx, account.GetAccountId(), policy, dynamicProviderInstanceConcurrencyLimit(provider, policy), holder)
}

func (p *dynamicIPSelector) dynamicIPEndpointCandidatesForAccount(ctx context.Context, account *proxyruntimev1.ProxyProviderAccount, accountIndex int, providerInstances []dynamicIPProviderInstance, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, sessionPolicy *proxyruntimev1.ProxySessionPolicy, filter dynamicIPCandidateFilter) []scoredDynamicIPEndpointCandidate {
	accountDynamicProviderID := runtimeSafeID(account.GetDynamicProviderId())
	out := []scoredDynamicIPEndpointCandidate{}
	for providerIndex, providerInstance := range providerInstances {
		if providerInstance.providerID != account.GetProviderId() {
			continue
		}
		if accountDynamicProviderID != "" && accountDynamicProviderID != providerInstance.dynamicProviderID {
			continue
		}
		if filter.dynamicProviderID != "" && filter.dynamicProviderID != providerInstance.dynamicProviderID {
			continue
		}
		if ok, err := p.providerAccountConcurrencyAvailable(ctx, account, providerInstance, sessionPolicy, filter.concurrencyHolder); err != nil || !ok {
			continue
		}
		for endpointIndex, endpoint := range providerInstance.endpoints {
			if strings.TrimSpace(endpoint.EndpointURL) == "" {
				continue
			}
			endpointID := firstNonEmpty(endpoint.ID, endpointIDFromURL(endpoint.EndpointURL))
			if filter.endpointID != "" && filter.endpointID != endpointID {
				continue
			}
			regions := p.endpointRegionCodes(ctx, endpoint, policy)
			candidate := &proxyruntimev1.ProxyDynamicIPEndpointCandidate{
				ProviderAccountId: account.GetAccountId(),
				ProviderId:        account.GetProviderId(),
				EndpointId:        endpointID,
				EndpointUrl:       endpoint.EndpointURL,
				GeoCodes:          cleanRegionCodes(regions),
				Protocol:          protocolEnum(p.endpointProtocolForProvider(account.GetProviderId(), endpoint)),
				Priority:          uint32(accountIndex*10000 + providerIndex*100 + endpointIndex),
				DynamicProviderId: providerInstance.dynamicProviderID,
			}
			scoredEndpoint := endpoint
			scoredEndpoint.ID = endpointID
			out = append(out, scoredDynamicIPEndpointCandidate{proto: candidate, endpoint: scoredEndpoint, score: endpointScore(regions, policy)})
		}
	}
	return out
}

func (p *dynamicIPSelector) endpointProtocolForProvider(providerID string, endpoint accountproxy.Gateway) string {
	if p != nil && p.accountProviders != nil {
		protocol, ok := p.accountProviders.GatewayProtocolForProvider(providerID, endpoint)
		if ok {
			return protocol
		}
	}
	return accountproxy.GatewayProtocol(endpoint, "socks5")
}

func chooseDynamicIPEndpointCandidate(candidates []scoredDynamicIPEndpointCandidate, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy, key string, attempt int) scoredDynamicIPEndpointCandidate {
	candidates = regionScopedDynamicIPEndpointCandidates(candidates, policy)
	groups := dynamicIPEndpointCandidateGroups(candidates, key)
	if len(groups) == 0 {
		return scoredDynamicIPEndpointCandidate{}
	}
	groupIndex := 0
	if len(groups) > 1 {
		if attempt > 1 {
			groupIndex = (attempt - 1) % len(groups)
		} else {
			groupIndex = int(hashModulo(key, uint32(len(groups))))
		}
	}
	return chooseDynamicIPEndpointWithinAccount(groups[groupIndex].candidates, key, attempt)
}

type dynamicIPEndpointCandidateGroup struct {
	providerAccountID string
	priority          uint32
	order             uint32
	candidates        []scoredDynamicIPEndpointCandidate
}

func dynamicIPEndpointCandidateGroups(candidates []scoredDynamicIPEndpointCandidate, key string) []dynamicIPEndpointCandidateGroup {
	byAccount := map[string]*dynamicIPEndpointCandidateGroup{}
	for _, candidate := range candidates {
		accountID := candidate.proto.GetProviderAccountId()
		if strings.TrimSpace(accountID) == "" {
			accountID = candidate.proto.GetProviderId()
		}
		group := byAccount[accountID]
		if group == nil {
			group = &dynamicIPEndpointCandidateGroup{
				providerAccountID: accountID,
				priority:          candidate.proto.GetPriority(),
				order:             hashModulo(firstNonEmpty(key, "proxy-runtime")+":"+accountID, 0),
			}
			byAccount[accountID] = group
		}
		if candidate.proto.GetPriority() < group.priority {
			group.priority = candidate.proto.GetPriority()
		}
		group.candidates = append(group.candidates, candidate)
	}
	out := make([]dynamicIPEndpointCandidateGroup, 0, len(byAccount))
	for _, group := range byAccount {
		out = append(out, *group)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].order != out[j].order {
			return out[i].order < out[j].order
		}
		if out[i].priority != out[j].priority {
			return out[i].priority < out[j].priority
		}
		return out[i].providerAccountID < out[j].providerAccountID
	})
	return out
}

func chooseDynamicIPEndpointWithinAccount(candidates []scoredDynamicIPEndpointCandidate, key string, attempt int) scoredDynamicIPEndpointCandidate {
	if len(candidates) == 0 {
		return scoredDynamicIPEndpointCandidate{}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].proto.GetPriority() < candidates[j].proto.GetPriority()
	})
	if len(candidates) > 1 {
		best := candidates[0].score
		count := 0
		for count < len(candidates) && candidates[count].score == best {
			count++
		}
		if attempt > 1 && count > 1 {
			return candidates[(attempt-1)%count]
		}
		if count > 1 {
			return candidates[int(hashModulo(key, uint32(count)))]
		}
	}
	return candidates[0]
}

func regionScopedDynamicIPEndpointCandidates(candidates []scoredDynamicIPEndpointCandidate, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) []scoredDynamicIPEndpointCandidate {
	if !hasRequestedRegion(policy) {
		return candidates
	}
	matched := make([]scoredDynamicIPEndpointCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if regionSpecificScore(candidate.proto.GetGeoCodes(), policy) > 0 {
			matched = append(matched, candidate)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	return candidates
}

func endpointScore(regions []string, policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) int {
	return 1000 + regionScore(regions, policy)
}
