package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func releaseProviderSession(ctx context.Context, providerClient provider.SessionProvider, session *proxyruntimev1.ProxySession) error {
	if providerClient == nil || session == nil || strings.TrimSpace(session.GetSessionId()) == "" {
		return nil
	}
	return providerClient.ReleaseSession(ctx, session)
}

func (c leaseCoordinator) releaseLeaseProviderSession(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	r := c.runtime
	if lease == nil || lease.GetSession() == nil || strings.TrimSpace(lease.GetProviderAccountId()) == "" {
		return nil
	}
	if statelessProviderSession(lease.GetSession()) {
		return nil
	}
	providerCfg, _, err := r.store.ProviderConfig(ctx, lease.GetProviderAccountId())
	if err != nil {
		return err
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		return err
	}
	providerCfg.Gateways = endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerCfg.ProviderID)
	providerClient, err := r.accountProviders.NewSessionProvider(providerCfg, BuildProviderHTTPClient(r.cfg))
	if err != nil {
		return err
	}
	return releaseProviderSession(ctx, providerClient, lease.GetSession())
}

func statelessProviderSession(session *proxyruntimev1.ProxySession) bool {
	switch strings.TrimSpace(session.GetLabels()["session_mode"]) {
	case "username_parameter", "provider_configured":
		return true
	default:
		return false
	}
}
