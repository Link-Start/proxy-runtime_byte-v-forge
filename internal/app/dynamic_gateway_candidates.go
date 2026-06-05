package app

import (
	"context"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/geox"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (p *dynamicGatewaySelector) dynamicGatewayCandidates(ctx context.Context, settings *runtimeSettingsFile, policy *proxyruntimev1.EgressRoutePolicy) ([]scoredGatewayCandidate, error) {
	accounts, err := p.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	busyProviderAccounts, err := p.store.ActiveProviderAccountIDs(ctx)
	if err != nil {
		p.logger.Warn("list active proxy provider accounts for gateway selection failed", "error", err)
		busyProviderAccounts = map[string]struct{}{}
	}
	gatewayMap := dynamicIPGatewayMap(settings)
	out := make([]scoredGatewayCandidate, 0)
	for accountIndex, account := range accounts {
		if account.GetStatus() != proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED || !account.GetCredentialConfigured() {
			continue
		}
		if !p.accountProviders.IsSupported(account.GetProviderId()) {
			continue
		}
		if requiresDynamicGeoTargeting(policy) && !p.accountProviders.SupportsRuntimeGeoTargeting(account.GetProviderId()) {
			continue
		}
		if _, busy := busyProviderAccounts[account.GetAccountId()]; busy {
			continue
		}
		for gatewayIndex, gateway := range gatewayMap[account.GetProviderId()] {
			if strings.TrimSpace(gateway.EndpointURL) == "" {
				continue
			}
			gatewayID := firstNonEmpty(gateway.ID, gatewayIDFromEndpointURL(gateway.EndpointURL))
			regions := p.gatewayRegionCodes(ctx, gateway, policy)
			candidate := &proxyruntimev1.ProxyDynamicGatewayCandidate{
				ProviderAccountId: account.GetAccountId(),
				ProviderId:        account.GetProviderId(),
				GatewayId:         gatewayID,
				DisplayName:       gateway.EndpointURL,
				RegionCodes:       cleanRegionCodes(regions),
				Protocol:          protocolEnum(p.gatewayProtocolForProvider(account.GetProviderId(), gateway)),
				Priority:          uint32(accountIndex*100 + gatewayIndex),
			}
			scoredGateway := gateway
			scoredGateway.ID = gatewayID
			out = append(out, scoredGatewayCandidate{proto: candidate, gateway: scoredGateway, score: gatewayScore(regions, policy)})
		}
	}
	return out, nil
}

func (p *dynamicGatewaySelector) gatewayProtocolForProvider(providerID string, gateway accountproxy.Gateway) string {
	protocol, ok := p.accountProviders.GatewayProtocolForProvider(providerID, gateway)
	if ok {
		return protocol
	}
	return accountproxy.GatewayProtocol(gateway, "socks5")
}

func chooseGatewayCandidate(candidates []scoredGatewayCandidate, policy *proxyruntimev1.EgressRoutePolicy, key string, attempt int) scoredGatewayCandidate {
	candidates = regionScopedGatewayCandidates(candidates, policy)
	groups := gatewayCandidateGroups(candidates, key)
	if len(groups) == 0 {
		return scoredGatewayCandidate{}
	}
	groupIndex := 0
	if len(groups) > 1 {
		if stableRouteStrategy(policy) {
			groupIndex = int(hashModulo(key, uint32(len(groups))))
		} else {
			if attempt < 1 {
				attempt = 1
			}
			groupIndex = (attempt - 1) % len(groups)
		}
	}
	return chooseGatewayWithinAccount(groups[groupIndex].candidates, policy, key, attempt)
}

type gatewayCandidateGroup struct {
	providerAccountID string
	priority          uint32
	order             uint32
	candidates        []scoredGatewayCandidate
}

func gatewayCandidateGroups(candidates []scoredGatewayCandidate, key string) []gatewayCandidateGroup {
	byAccount := map[string]*gatewayCandidateGroup{}
	for _, candidate := range candidates {
		accountID := candidate.proto.GetProviderAccountId()
		if strings.TrimSpace(accountID) == "" {
			accountID = candidate.proto.GetProviderId()
		}
		group := byAccount[accountID]
		if group == nil {
			group = &gatewayCandidateGroup{
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
	out := make([]gatewayCandidateGroup, 0, len(byAccount))
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

func chooseGatewayWithinAccount(candidates []scoredGatewayCandidate, policy *proxyruntimev1.EgressRoutePolicy, key string, attempt int) scoredGatewayCandidate {
	if len(candidates) == 0 {
		return scoredGatewayCandidate{}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].proto.GetPriority() < candidates[j].proto.GetPriority()
	})
	if stableRouteStrategy(policy) && len(candidates) > 1 {
		return candidates[int(hashModulo(key, uint32(len(candidates))))]
	}
	if attempt > 1 && len(candidates) > 1 {
		best := candidates[0].score
		count := 0
		for count < len(candidates) && candidates[count].score == best {
			count++
		}
		if count > 1 {
			return candidates[(attempt-1)%count]
		}
	}
	return candidates[0]
}

func regionScopedGatewayCandidates(candidates []scoredGatewayCandidate, policy *proxyruntimev1.EgressRoutePolicy) []scoredGatewayCandidate {
	if !hasRequestedRegion(policy) {
		return candidates
	}
	matched := make([]scoredGatewayCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if regionSpecificScore(candidate.proto.GetRegionCodes(), policy) > 0 {
			matched = append(matched, candidate)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	return candidates
}

func requiresDynamicGeoTargeting(policy *proxyruntimev1.EgressRoutePolicy) bool {
	return geox.NormalizeCountryAlpha2(policy.GetCountryCode()) != ""
}

func gatewayScore(regions []string, policy *proxyruntimev1.EgressRoutePolicy) int {
	return 1000 + regionScore(regions, policy)
}
