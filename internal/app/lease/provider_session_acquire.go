package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

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
	ErrorKind         ProviderSessionErrorKind
}

func AcquireProviderSession(ctx context.Context, input ProviderSessionAcquireInput) (ProviderSessionAcquireResult, error) {
	providerCfg, accountID, err := ProviderConfigForGateway(ctx, input.Store, input.ProviderAccountID, input.Gateway)
	if err != nil {
		return ProviderSessionAcquireResult{}, err
	}
	providerClient, err := NewSessionProvider(input.Factory, providerCfg)
	result := ProviderSessionAcquireResult{ProviderAccountID: accountID, ProviderClient: providerClient}
	if err != nil {
		result.ErrorKind = ProviderSessionFactoryError
		return result, err
	}
	session, nodes, err := CreateAndFetchProviderSession(ctx, providerClient, input.Request, input.SelectionPlan, input.ConcurrencyHolder)
	result.Session = session
	result.Nodes = nodes
	result.ErrorKind = ClassifyProviderSessionError(session, err)
	return result, err
}

func NewSessionProvider(factory SessionProviderFactory, providerCfg accountproxy.Config) (SessionProvider, error) {
	if factory == nil {
		return nil, ErrSessionProviderFactoryRequired
	}
	return factory.NewSessionProvider(providerCfg)
}
