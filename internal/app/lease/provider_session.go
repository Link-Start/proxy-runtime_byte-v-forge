package lease

import (
	"context"
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

var ErrSessionProviderRequired = errors.New("provider session provider is required")

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

func StatelessProviderSession(session *proxyruntimev1.ProxySession) bool {
	switch strings.TrimSpace(session.GetLabels()["session_mode"]) {
	case "username_parameter", "provider_configured":
		return true
	default:
		return false
	}
}
