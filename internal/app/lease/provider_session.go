package lease

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

var (
	ErrSessionProviderRequired = errors.New("provider session provider is required")
	ErrProviderSessionFactory  = errors.New("provider session factory failed")
	ErrProviderSessionCreate   = errors.New("provider session create failed")
	ErrProviderSessionFetch    = errors.New("provider session fetch failed")
)

type ProviderSessionErrorKind int

const (
	ProviderSessionNoError ProviderSessionErrorKind = iota
	ProviderSessionFactoryError
	ProviderSessionCreateError
	ProviderSessionFetchError
)

func ClassifyProviderSessionError(session *proxyruntimev1.ProxySession, err error) ProviderSessionErrorKind {
	if err == nil {
		return ProviderSessionNoError
	}
	if session == nil {
		return ProviderSessionCreateError
	}
	return ProviderSessionFetchError
}

func ProviderSessionStageError(kind ProviderSessionErrorKind, err error) error {
	if err == nil {
		return nil
	}
	switch kind {
	case ProviderSessionFactoryError:
		return fmt.Errorf("%w: %w", ErrProviderSessionFactory, err)
	case ProviderSessionCreateError:
		return fmt.Errorf("%w: %w", ErrProviderSessionCreate, err)
	case ProviderSessionFetchError:
		return fmt.Errorf("%w: %w", ErrProviderSessionFetch, err)
	default:
		return err
	}
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
