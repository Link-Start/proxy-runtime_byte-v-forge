package dynamic

import (
	"context"
	"errors"
	"fmt"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
)

type ScoredEndpointCandidate struct {
	Proto    *proxyruntimev1.ProxyDynamicIPEndpointCandidate
	Endpoint accountproxy.Gateway
	score    int
}

type IPSelector struct {
	store            dynamicIPSelectionStore
	loadSettingsFn   func(context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
	accountProviders dynamicIPSelectionProviderRegistry
	concurrency      leaseapp.ProviderAccountConcurrencyLimiter
	logger           dynamicIPSelectionLogger
	lookupIPGeo      func(context.Context, string) (proxycheck.ExitGeo, error)
	clock            clock.Clock
}

type IPSelectorDependencies struct {
	Store            dynamicIPSelectionStore
	LoadSettings     func(context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error)
	AccountProviders dynamicIPSelectionProviderRegistry
	Concurrency      leaseapp.ProviderAccountConcurrencyLimiter
	Logger           dynamicIPSelectionLogger
	LookupIPGeo      func(context.Context, string) (proxycheck.ExitGeo, error)
	Clock            clock.Clock
}

type dynamicIPSelectionStore interface {
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	RecentLeaseFacts(context.Context, time.Time, int) ([]*proxyruntimev1.ProxyDynamicLease, error)
}

type dynamicIPSelectionProviderRegistry interface {
	IsSupported(string) bool
	GatewayProtocolForProvider(string, accountproxy.Gateway) (string, bool)
}

type dynamicIPSelectionLogger interface {
	Warn(string, ...any)
}

func NewIPSelector(deps IPSelectorDependencies) *IPSelector {
	return &IPSelector{
		store:            deps.Store,
		loadSettingsFn:   deps.LoadSettings,
		accountProviders: deps.AccountProviders,
		concurrency:      deps.Concurrency,
		logger:           deps.Logger,
		lookupIPGeo:      deps.LookupIPGeo,
		clock:            deps.Clock,
	}
}

func (p *IPSelector) SelectDynamicIPEndpoint(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (leaseapp.DynamicIPSelection, error) {
	settings, err := p.loadSettings(ctx)
	if err != nil {
		return leaseapp.DynamicIPSelection{}, err
	}
	policy := leaseapp.NormalizeDynamicIPSelectionPolicy(req)
	endpoints, err := p.dynamicIPEndpointCandidates(ctx, settings, policy, req.GetPolicy())
	if err != nil {
		return leaseapp.DynamicIPSelection{}, err
	}
	if len(endpoints) == 0 {
		return leaseapp.DynamicIPSelection{}, errors.New("no dynamic IP endpoint candidate")
	}
	attempt := leaseapp.DynamicIPSelectionAttempt(req)
	selectedEndpoint := ChooseEndpointCandidate(endpoints, policy, leaseapp.DynamicIPSelectionKey(req), attempt)
	reasons := []string{
		fmt.Sprintf("dynamic_ip_endpoint=%s/%s/%s/%s", selectedEndpoint.Proto.GetProviderAccountId(), selectedEndpoint.Proto.GetProviderId(), selectedEndpoint.Proto.GetDynamicProviderId(), selectedEndpoint.Proto.GetEndpointId()),
	}
	selectionID := "selection-" + appcore.ShortHash(req.GetAccountId()+":"+policy.GetPurpose())
	plan := &proxyruntimev1.ProxyDynamicIPSelectionPlan{
		SelectionId:      selectionID,
		Policy:           dynamicIPSelectionPlanPolicy(policy),
		SelectedEndpoint: selectedEndpoint.Proto,
		SelectionReasons: reasons,
		SelectedAt:       timestamppb.New(p.clock.Now().UTC()),
	}
	return leaseapp.DynamicIPSelection{Plan: plan, Endpoint: selectedEndpoint.Endpoint}, nil
}

func (p *IPSelector) loadSettings(ctx context.Context) (*proxyruntimev1.ProxyRuntimePersistentSettings, error) {
	if p == nil || p.loadSettingsFn == nil {
		return nil, appcore.InternalError("dynamic IP selection settings repository is not configured", nil)
	}
	return p.loadSettingsFn(ctx)
}

func (p *IPSelector) providerSupported(providerID string) bool {
	return p != nil && p.accountProviders != nil && p.accountProviders.IsSupported(providerID)
}

func (p *IPSelector) warn(message string, args ...any) {
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
