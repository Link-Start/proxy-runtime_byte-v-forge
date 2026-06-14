package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) saveFailedAcquireLeaseFact(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, session *proxyruntimev1.ProxySession, egress *proxyruntimev1.ProxyEndpoint, listener *proxyruntimev1.EgressListener, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan, message string) {
	lease, err := leaseapp.SaveFailedAcquireFact(ctx, c.deps.store, c.deps.ids, c.deps.clock, req, providerAccountID, session, egress, listener, plan, message)
	if err != nil && lease != nil {
		c.warn("save failed proxy lease fact failed", "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
	}
}
