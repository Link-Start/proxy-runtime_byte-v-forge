package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func MarkFailedAcquireBeforeRouteCleanup(ctx context.Context, providerClient SessionProvider, session *proxyruntimev1.ProxySession) error {
	providerPending, err := failedAcquireProviderCleanupPending(ctx, providerClient, session)
	MarkFailedAcquireCleanupPending(session, false, providerPending)
	return err
}

func MarkFailedAcquireAfterRouteCleanup(ctx context.Context, dataPlane DataPlaneApplier, route SessionRoute, providerClient SessionProvider, session *proxyruntimev1.ProxySession) error {
	routePending := DeleteSessionRoute(ctx, dataPlane, route) != nil
	providerPending, err := failedAcquireProviderCleanupPending(ctx, providerClient, session)
	MarkFailedAcquireCleanupPending(session, routePending, providerPending)
	return err
}

func failedAcquireProviderCleanupPending(ctx context.Context, providerClient SessionProvider, session *proxyruntimev1.ProxySession) (bool, error) {
	err := ReleaseProviderSession(ctx, providerClient, session)
	return err != nil, err
}
