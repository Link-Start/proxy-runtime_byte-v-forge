package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) warn(message string, args ...any) {
	if c.deps.logger != nil {
		c.deps.logger.Warn(message, args...)
	}
}

func (c leaseCoordinator) warnFinalConcurrencyReleaseFailed(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) {
	_ = ctx
	c.warn("release provider account concurrency slot failed", "lease_id", lease.GetLeaseId(), "provider_account_id", lease.GetProviderAccountId())
}

func (c leaseCoordinator) now() time.Time {
	if c.deps.clock != nil {
		return c.deps.clock.Now()
	}
	return time.Now()
}

func (c leaseCoordinator) clearExitCheckCache() {
	if c.deps.exitCheckCache != nil {
		c.deps.exitCheckCache.clear()
	}
}

func (c leaseCoordinator) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if c.deps.closeInUserConnections != nil {
		c.deps.closeInUserConnections(ctx, usernames)
	}
}

func (c leaseCoordinator) providerSessionGatewaysResolver(lease *proxyruntimev1.ProxyDynamicLease) func(context.Context, string) ([]accountproxy.Gateway, error) {
	return func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
		settings, err := c.deps.settings.load(ctx)
		if err != nil {
			return nil, err
		}
		return endpointsForDynamicIPSelection(settings, lease.GetSelectionPlan(), providerID), nil
	}
}
