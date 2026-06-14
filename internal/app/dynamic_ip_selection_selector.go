package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
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
	store            dynamicIPSelectionStore
	settings         dynamicIPSelectionSettings
	accountProviders dynamicIPSelectionProviderRegistry
	concurrency      providerAccountConcurrencyLimiter
	logger           dynamicIPSelectionLogger
	lookupIPGeo      func(context.Context, string) (proxyExitGeo, error)
}

type dynamicIPSelectorDependencies struct {
	Store            dynamicIPSelectionStore
	Settings         dynamicIPSelectionSettings
	AccountProviders dynamicIPSelectionProviderRegistry
	Concurrency      providerAccountConcurrencyLimiter
	Logger           dynamicIPSelectionLogger
	LookupIPGeo      func(context.Context, string) (proxyExitGeo, error)
}

type dynamicIPSelectionStore interface {
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	RecentLeaseFacts(context.Context, time.Time, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
}

type dynamicIPSelectionSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type dynamicIPSelectionProviderRegistry interface {
	IsSupported(string) bool
	GatewayProtocolForProvider(string, accountproxy.Gateway) (string, bool)
}

type dynamicIPSelectionLogger interface {
	Warn(string, ...any)
}

func newDynamicIPSelector(deps dynamicIPSelectorDependencies) *dynamicIPSelector {
	return &dynamicIPSelector{
		store:            deps.Store,
		settings:         deps.Settings,
		accountProviders: deps.AccountProviders,
		concurrency:      deps.Concurrency,
		logger:           deps.Logger,
		lookupIPGeo:      deps.LookupIPGeo,
	}
}

func (p *dynamicIPSelector) selectDynamicIPEndpoint(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (dynamicIPSelection, error) {
	settings, err := p.loadSettings(ctx)
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

func (p *dynamicIPSelector) loadSettings(ctx context.Context) (*runtimeSettingsFile, error) {
	if p == nil || p.settings == nil {
		return nil, internalError("dynamic IP selection settings repository is not configured", nil)
	}
	return p.settings.load(ctx)
}

func (p *dynamicIPSelector) providerSupported(providerID string) bool {
	return p != nil && p.accountProviders != nil && p.accountProviders.IsSupported(providerID)
}

func (p *dynamicIPSelector) warn(message string, args ...any) {
	if p != nil && p.logger != nil {
		p.logger.Warn(message, args...)
	}
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
