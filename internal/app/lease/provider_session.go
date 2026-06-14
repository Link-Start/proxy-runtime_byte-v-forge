package lease

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
