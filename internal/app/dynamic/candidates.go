package dynamic

import (
	"context"
	"sort"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func (p *IPSelector) dynamicIPEndpointCandidates(ctx context.Context, settings *proxygatewayv1.ProxyGatewayPersistentSettings, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy, sessionPolicy *proxygatewayv1.ProxySessionPolicy) ([]ScoredEndpointCandidate, error) {
	if p == nil || p.store == nil {
		return nil, appcore.InternalError("dynamic IP selection store is not configured", nil)
	}
	accounts, err := p.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	providerInstances := ProviderInstances(settings)
	filter := dynamicIPCandidateFilterFromPolicy(sessionPolicy)
	out := make([]ScoredEndpointCandidate, 0)
	for accountIndex, account := range accounts {
		if account.GetStatus() != proxygatewayv1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || !account.GetCredentialConfigured() {
			continue
		}
		if !p.providerSupported(account.GetProviderId()) {
			continue
		}
		out = append(out, p.DynamicIPEndpointCandidatesForAccount(ctx, account, accountIndex, providerInstances, policy, sessionPolicy, filter)...)
	}
	ApplyEndpointHealthScores(out, p.DynamicIPEndpointHealthScores(ctx))
	return out, nil
}

type IPCandidateFilter struct {
	dynamicProviderID string
	endpointID        string
	ConcurrencyHolder string
}

func dynamicIPCandidateFilterFromPolicy(policy *proxygatewayv1.ProxySessionPolicy) IPCandidateFilter {
	labels := policy.GetLabels()
	return IPCandidateFilter{
		dynamicProviderID: appcore.RuntimeSafeID(labels["dynamic_provider_id"]),
		endpointID:        strings.TrimSpace(labels["dynamic_ip_endpoint_id"]),
	}
}

func (p *IPSelector) providerAccountConcurrencyAvailable(ctx context.Context, account *proxygatewayv1.ProxyProviderAccount, provider ProviderInstance, policy *proxygatewayv1.ProxySessionPolicy, holder string) (bool, error) {
	if p.concurrency == nil {
		return true, nil
	}
	return p.concurrency.Available(ctx, account.GetAccountId(), policy, ProviderInstanceConcurrencyLimit(provider, policy), holder)
}

func (p *IPSelector) DynamicIPEndpointCandidatesForAccount(ctx context.Context, account *proxygatewayv1.ProxyProviderAccount, accountIndex int, providerInstances []ProviderInstance, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy, sessionPolicy *proxygatewayv1.ProxySessionPolicy, filter IPCandidateFilter) []ScoredEndpointCandidate {
	accountDynamicProviderID := appcore.RuntimeSafeID(account.GetDynamicProviderId())
	out := []ScoredEndpointCandidate{}
	for providerIndex, providerInstance := range providerInstances {
		if providerInstance.ProviderID != account.GetProviderId() {
			continue
		}
		if accountDynamicProviderID != "" && accountDynamicProviderID != providerInstance.DynamicProviderID {
			continue
		}
		if filter.dynamicProviderID != "" && filter.dynamicProviderID != providerInstance.DynamicProviderID {
			continue
		}
		if ok, err := p.providerAccountConcurrencyAvailable(ctx, account, providerInstance, sessionPolicy, filter.ConcurrencyHolder); err != nil || !ok {
			continue
		}
		for endpointIndex, endpoint := range providerInstance.Endpoints {
			if strings.TrimSpace(endpoint.EndpointURL) == "" {
				continue
			}
			endpointID := appcore.FirstNonEmpty(endpoint.ID, kernel.EndpointIDFromURL(endpoint.EndpointURL))
			if filter.endpointID != "" && filter.endpointID != endpointID {
				continue
			}
			regions := p.endpointRegionCodes(ctx, endpoint, policy)
			candidate := &proxygatewayv1.ProxyDynamicIPEndpointCandidate{
				ProviderAccountId: account.GetAccountId(),
				ProviderId:        account.GetProviderId(),
				EndpointId:        endpointID,
				EndpointUrl:       endpoint.EndpointURL,
				GeoCodes:          cleanRegionCodes(regions),
				Protocol:          protocolEnum(p.endpointProtocolForProvider(account.GetProviderId(), endpoint)),
				Priority:          uint32(accountIndex*10000 + providerIndex*100 + endpointIndex),
				DynamicProviderId: providerInstance.DynamicProviderID,
			}
			scoredEndpoint := endpoint
			scoredEndpoint.ID = endpointID
			out = append(out, ScoredEndpointCandidate{Proto: candidate, Endpoint: scoredEndpoint, score: endpointScore(regions, policy)})
		}
	}
	return out
}

func (p *IPSelector) endpointProtocolForProvider(providerID string, endpoint accountproxy.Gateway) string {
	if p != nil && p.accountProviders != nil {
		protocol, ok := p.accountProviders.GatewayProtocolForProvider(providerID, endpoint)
		if ok {
			return protocol
		}
	}
	return accountproxy.GatewayProtocol(endpoint, "socks5")
}

func ChooseEndpointCandidate(candidates []ScoredEndpointCandidate, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy, key string, attempt int) ScoredEndpointCandidate {
	candidates = regionScopedDynamicIPEndpointCandidates(candidates, policy)
	groups := dynamicIPEndpointCandidateGroups(candidates, key)
	if len(groups) == 0 {
		return ScoredEndpointCandidate{}
	}
	groupIndex := 0
	if len(groups) > 1 {
		if attempt > 1 {
			groupIndex = (attempt - 1) % len(groups)
		} else {
			groupIndex = int(appcore.HashModulo(key, uint32(len(groups))))
		}
	}
	return chooseDynamicIPEndpointWithinAccount(groups[groupIndex].candidates, key, attempt)
}

type dynamicIPEndpointCandidateGroup struct {
	providerAccountID string
	priority          uint32
	order             uint32
	candidates        []ScoredEndpointCandidate
}

func dynamicIPEndpointCandidateGroups(candidates []ScoredEndpointCandidate, key string) []dynamicIPEndpointCandidateGroup {
	byAccount := map[string]*dynamicIPEndpointCandidateGroup{}
	for _, candidate := range candidates {
		accountID := candidate.Proto.GetProviderAccountId()
		if strings.TrimSpace(accountID) == "" {
			accountID = candidate.Proto.GetProviderId()
		}
		group := byAccount[accountID]
		if group == nil {
			group = &dynamicIPEndpointCandidateGroup{
				providerAccountID: accountID,
				priority:          candidate.Proto.GetPriority(),
				order:             appcore.HashModulo(appcore.FirstNonEmpty(key, "proxy-gateway")+":"+accountID, 0),
			}
			byAccount[accountID] = group
		}
		if candidate.Proto.GetPriority() < group.priority {
			group.priority = candidate.Proto.GetPriority()
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

func chooseDynamicIPEndpointWithinAccount(candidates []ScoredEndpointCandidate, key string, attempt int) ScoredEndpointCandidate {
	if len(candidates) == 0 {
		return ScoredEndpointCandidate{}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].Proto.GetPriority() < candidates[j].Proto.GetPriority()
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
			return candidates[int(appcore.HashModulo(key, uint32(count)))]
		}
	}
	return candidates[0]
}

func regionScopedDynamicIPEndpointCandidates(candidates []ScoredEndpointCandidate, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) []ScoredEndpointCandidate {
	if !hasRequestedRegion(policy) {
		return candidates
	}
	matched := make([]ScoredEndpointCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if regionSpecificScore(candidate.Proto.GetGeoCodes(), policy) > 0 {
			matched = append(matched, candidate)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	return candidates
}

func endpointScore(regions []string, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) int {
	return 1000 + regionScore(regions, policy)
}
