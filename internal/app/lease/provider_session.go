package lease

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

var (
	ErrSessionProviderRequired = errors.New("provider session provider is required")
	ErrProviderSessionFactory  = errors.New("provider session factory failed")
	ErrProviderSessionCreate   = errors.New("provider session create failed")
	ErrProviderSessionFetch    = errors.New("provider session fetch failed")
)

func WrapProviderSessionFactoryFailure(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrProviderSessionFactory, err)
}

func WrapProviderSessionCreateFetchFailure(session *proxyruntimev1.ProxySession, err error) error {
	if err == nil {
		return nil
	}
	if session == nil {
		return fmt.Errorf("%w: %w", ErrProviderSessionCreate, err)
	}
	return fmt.Errorf("%w: %w", ErrProviderSessionFetch, err)
}

func CreateAndFetchProviderSession(ctx context.Context, providerClient SessionProvider, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, concurrencyHolder string) (*proxyruntimev1.ProxySession, []provider.Node, error) {
	session, err := CreateProviderSession(ctx, providerClient, req, selectionPlan, concurrencyHolder)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := FetchProviderSession(ctx, providerClient, session)
	if err != nil {
		return session, nil, err
	}
	return session, nodes, nil
}

func CreateProviderSession(ctx context.Context, providerClient SessionProvider, req *proxyruntimev1.AcquireProxyLeaseRequest, selectionPlan *proxyruntimev1.ProxyDynamicIPSelectionPlan, concurrencyHolder string) (*proxyruntimev1.ProxySession, error) {
	if providerClient == nil {
		return nil, ErrSessionProviderRequired
	}
	ApplyProviderSessionRequestLabels(req, selectionPlan, concurrencyHolder)
	return providerClient.CreateSession(ctx, req)
}

func FetchProviderSession(ctx context.Context, providerClient SessionProvider, session *proxyruntimev1.ProxySession) ([]provider.Node, error) {
	if providerClient == nil {
		return nil, ErrSessionProviderRequired
	}
	return providerClient.FetchSession(ctx, session)
}

func ReleaseProviderSession(ctx context.Context, providerClient SessionProvider, session *proxyruntimev1.ProxySession) error {
	if providerClient == nil || session == nil || strings.TrimSpace(session.GetSessionId()) == "" {
		return nil
	}
	return providerClient.ReleaseSession(ctx, session)
}

func ProviderName(providerClient SessionProvider) string {
	if providerClient == nil {
		return ""
	}
	return providerClient.Name()
}

func StatelessProviderSession(session *proxyruntimev1.ProxySession) bool {
	switch strings.TrimSpace(session.GetLabels()["session_mode"]) {
	case "username_parameter", "provider_configured":
		return true
	default:
		return false
	}
}

var ErrSessionProviderFactoryRequired = errors.New("provider session factory is required")

type ProviderSessionAcquireInput struct {
	Store             OrchestrationStore
	Factory           SessionProviderFactory
	ProviderAccountID string
	Gateway           accountproxy.Gateway
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ConcurrencyHolder string
}

type ProviderSessionAcquireResult struct {
	ProviderAccountID string
	ProviderClient    SessionProvider
	Session           *proxyruntimev1.ProxySession
	Nodes             []provider.Node
}

func AcquireProviderSession(ctx context.Context, input ProviderSessionAcquireInput) (ProviderSessionAcquireResult, error) {
	providerCfg, accountID, err := ProviderConfigForGateway(ctx, input.Store, input.ProviderAccountID, input.Gateway)
	if err != nil {
		return ProviderSessionAcquireResult{}, err
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	result := ProviderSessionAcquireResult{ProviderAccountID: accountID, ProviderClient: providerClient}
	if err != nil {
		return result, WrapProviderSessionFactoryFailure(err)
	}
	session, nodes, err := CreateAndFetchProviderSession(ctx, providerClient, input.Request, input.SelectionPlan, input.ConcurrencyHolder)
	result.Session = session
	result.Nodes = nodes
	return result, WrapProviderSessionCreateFetchFailure(session, err)
}

func NewSessionProvider(factory SessionProviderFactory, providerCfg accountproxy.Config) (SessionProvider, error) {
	if factory == nil {
		return nil, ErrSessionProviderFactoryRequired
	}
	return factory.NewSessionProvider(providerCfg)
}

type ProviderSessionFetchInput struct {
	Factory         SessionProviderFactory
	Lease           *proxyruntimev1.ProxyDynamicLease
	ProviderConfig  accountproxy.Config
	ResolveGateways ProviderSessionGatewaysResolver
}

func FetchLeaseProviderSession(ctx context.Context, input ProviderSessionFetchInput) ([]provider.Node, error) {
	providerCfg := input.ProviderConfig
	if input.ResolveGateways != nil {
		gateways, err := input.ResolveGateways(ctx, providerCfg.ProviderID)
		if err != nil {
			return nil, err
		}
		providerCfg.Gateways = gateways
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	if err != nil {
		return nil, err
	}
	return FetchProviderSession(ctx, providerClient, input.Lease.GetSession())
}

type ProviderSessionGatewaysResolver func(context.Context, string) ([]accountproxy.Gateway, error)

type ProviderSessionReleaseInput struct {
	Store           OrchestrationStore
	Factory         SessionProviderFactory
	Lease           *proxyruntimev1.ProxyDynamicLease
	ResolveGateways ProviderSessionGatewaysResolver
	RecordFailure   ProviderSessionReleaseFailureRecorder
}

type ProviderSessionReleaseFailureRecorder func(context.Context, *proxyruntimev1.ProxyDynamicLease, error) error

func ReleaseLeaseProviderSession(ctx context.Context, input ProviderSessionReleaseInput) error {
	if !NeedsProviderSessionRelease(input.Lease) {
		return nil
	}
	providerCfg, _, err := ProviderConfigForLease(ctx, input.Store, input.Lease)
	if err != nil {
		return err
	}
	if input.ResolveGateways != nil {
		gateways, err := input.ResolveGateways(ctx, providerCfg.ProviderID)
		if err != nil {
			return err
		}
		providerCfg.Gateways = gateways
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	if err != nil {
		return err
	}
	return ReleaseProviderSession(ctx, providerClient, input.Lease.GetSession())
}

func NeedsProviderSessionRelease(lease *proxyruntimev1.ProxyDynamicLease) bool {
	if lease == nil || lease.GetSession() == nil || strings.TrimSpace(lease.GetProviderAccountId()) == "" {
		return false
	}
	return !StatelessProviderSession(lease.GetSession())
}

func ReleaseLeaseProviderSessionWithLock(ctx context.Context, locks LockManager, input ProviderSessionReleaseInput) error {
	return WithProviderAccountLock(ctx, locks, input.Lease.GetProviderAccountId(), func(ctx context.Context) error {
		err := ReleaseLeaseProviderSession(ctx, input)
		if err != nil && input.RecordFailure != nil {
			_ = input.RecordFailure(ctx, input.Lease, err)
		}
		return err
	})
}
