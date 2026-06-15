package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
