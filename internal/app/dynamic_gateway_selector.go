package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type dynamicGatewaySelection struct {
	plan    *proxyruntimev1.EgressRoutePlan
	gateway accountproxy.Gateway
}

type scoredGatewayCandidate struct {
	proto   *proxyruntimev1.ProxyDynamicGatewayCandidate
	gateway accountproxy.Gateway
	score   int
}

type dynamicGatewaySelector struct {
	store            *PostgresStore
	settings         *runtimeSettingsStore
	accountProviders *providerregistry.Registry
	logger           *slog.Logger
	lookupIPGeo      func(context.Context, string) (proxyExitGeo, error)
}

func newDynamicGatewaySelector(runtime *Runtime) *dynamicGatewaySelector {
	return &dynamicGatewaySelector{
		store:            runtime.store,
		settings:         runtime.settings,
		accountProviders: runtime.accountProviders,
		logger:           runtime.logger,
		lookupIPGeo:      runtime.lookupIPGeo,
	}
}

func (p *dynamicGatewaySelector) selectDynamicGateway(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (dynamicGatewaySelection, error) {
	settings, err := p.settings.load(ctx)
	if err != nil {
		return dynamicGatewaySelection{}, err
	}
	policy := normalizeDynamicGatewayPolicy(req)
	gateways, err := p.dynamicGatewayCandidates(ctx, settings, policy)
	if err != nil {
		return dynamicGatewaySelection{}, err
	}
	if len(gateways) == 0 {
		return dynamicGatewaySelection{}, errors.New("no dynamic IP gateway candidate")
	}
	attempt := gatewayAttempt(req)
	selectedGateway := chooseGatewayCandidate(gateways, policy, gatewaySelectionKey(req), attempt)
	reasons := []string{
		fmt.Sprintf("dynamic_gateway=%s/%s/%s", selectedGateway.proto.GetProviderAccountId(), selectedGateway.proto.GetProviderId(), selectedGateway.proto.GetGatewayId()),
		"route=direct_dynamic_gateway",
	}
	routeID := "route-" + shortHash(req.GetAccountId()+":"+policy.GetPurpose())
	plan := &proxyruntimev1.EgressRoutePlan{
		RouteId:          routeID,
		Policy:           dynamicIPPlanPolicy(policy),
		DynamicGateway:   selectedGateway.proto,
		SelectionReasons: reasons,
		PlannedAt:        timestamppb.New(time.Now().UTC()),
		Route: &proxyruntimev1.EgressRoute{
			RouteId: routeID,
			Hops:    p.dynamicGatewayRouteHops(ctx, selectedGateway),
		},
	}
	return dynamicGatewaySelection{plan: plan, gateway: selectedGateway.gateway}, nil
}

func dynamicIPPlanPolicy(policy *proxyruntimev1.EgressRoutePolicy) *proxyruntimev1.EgressRoutePolicy {
	out := cloneRoutePolicy(policy)
	out.AllowDirectDynamicGateway = true
	return out
}

func cloneRoutePolicy(policy *proxyruntimev1.EgressRoutePolicy) *proxyruntimev1.EgressRoutePolicy {
	if policy == nil {
		return &proxyruntimev1.EgressRoutePolicy{}
	}
	return &proxyruntimev1.EgressRoutePolicy{
		CountryCode:               policy.GetCountryCode(),
		Region:                    policy.GetRegion(),
		Purpose:                   policy.GetPurpose(),
		Strategy:                  policy.GetStrategy(),
		MaxAttempts:               policy.GetMaxAttempts(),
		RequireDynamicExit:        policy.GetRequireDynamicExit(),
		AllowDirectDynamicGateway: policy.GetAllowDirectDynamicGateway(),
	}
}
