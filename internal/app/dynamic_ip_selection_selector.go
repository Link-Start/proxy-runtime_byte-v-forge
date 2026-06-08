package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type dynamicIPSelection struct {
	plan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	endpoint accountproxy.Gateway
}

type scoredDynamicIPEndpointCandidate struct {
	proto    *proxyruntimev1.ProxyDynamicIPEndpointCandidate
	endpoint accountproxy.Gateway
	score    int
}

type dynamicIPSelector struct {
	store            controlStore
	settings         *runtimeSettingsStore
	accountProviders *providerregistry.Registry
	concurrency      providerAccountConcurrencyLimiter
	logger           *slog.Logger
	lookupIPGeo      func(context.Context, string) (proxyExitGeo, error)
}

func newDynamicIPSelector(runtime *Runtime) *dynamicIPSelector {
	return &dynamicIPSelector{
		store:            runtime.store,
		settings:         runtime.settings,
		accountProviders: runtime.accountProviders,
		concurrency:      runtime.providerConcurrency,
		logger:           runtime.logger,
		lookupIPGeo:      runtime.lookupIPGeo,
	}
}

func (p *dynamicIPSelector) selectDynamicIPEndpoint(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (dynamicIPSelection, error) {
	settings, err := p.settings.load(ctx)
	if err != nil {
		return dynamicIPSelection{}, err
	}
	policy := normalizeDynamicIPSelectionPolicy(req)
	endpoints, err := p.dynamicIPEndpointCandidates(ctx, settings, policy, req.GetPolicy())
	if err != nil {
		return dynamicIPSelection{}, err
	}
	if len(endpoints) == 0 {
		return dynamicIPSelection{}, errors.New("no dynamic IP endpoint candidate")
	}
	attempt := dynamicIPSelectionAttempt(req)
	selectedEndpoint := chooseDynamicIPEndpointCandidate(endpoints, policy, dynamicIPSelectionKey(req), attempt)
	reasons := []string{
		fmt.Sprintf("dynamic_ip_endpoint=%s/%s/%s/%s", selectedEndpoint.proto.GetProviderAccountId(), selectedEndpoint.proto.GetProviderId(), selectedEndpoint.proto.GetDynamicProviderId(), selectedEndpoint.proto.GetEndpointId()),
	}
	selectionID := "selection-" + shortHash(req.GetAccountId()+":"+policy.GetPurpose())
	plan := &proxyruntimev1.ProxyDynamicIPSelectionPlan{
		SelectionId:      selectionID,
		Policy:           dynamicIPSelectionPlanPolicy(policy),
		SelectedEndpoint: selectedEndpoint.proto,
		SelectionReasons: reasons,
		SelectedAt:       timestamppb.New(time.Now().UTC()),
	}
	return dynamicIPSelection{plan: plan, endpoint: selectedEndpoint.endpoint}, nil
}

func dynamicIPSelectionPlanPolicy(policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) *proxyruntimev1.ProxyDynamicIPSelectionPolicy {
	return cloneDynamicIPSelectionPolicy(policy)
}

func cloneDynamicIPSelectionPolicy(policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) *proxyruntimev1.ProxyDynamicIPSelectionPolicy {
	if policy == nil {
		return &proxyruntimev1.ProxyDynamicIPSelectionPolicy{}
	}
	return &proxyruntimev1.ProxyDynamicIPSelectionPolicy{
		CountryCode: policy.GetCountryCode(),
		Region:      policy.GetRegion(),
		Purpose:     policy.GetPurpose(),
		MaxAttempts: policy.GetMaxAttempts(),
	}
}
