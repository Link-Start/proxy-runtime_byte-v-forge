package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type leaseAcquireFailure struct {
	coordinator       leaseCoordinator
	ctx               context.Context
	req               *proxyruntimev1.AcquireProxyLeaseRequest
	providerAccountID string
	providerClient    leaseapp.SessionProvider
	session           *proxyruntimev1.ProxySession
	listener          *proxyruntimev1.EgressListener
	egress            *proxyruntimev1.ProxyEndpoint
	plan              *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

func newLeaseAcquireFailure(coordinator leaseCoordinator, ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, providerClient leaseapp.SessionProvider, session *proxyruntimev1.ProxySession, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan) *leaseAcquireFailure {
	return &leaseAcquireFailure{coordinator: coordinator, ctx: ctx, req: req, providerAccountID: providerAccountID, providerClient: providerClient, session: session, plan: plan}
}

func (f *leaseAcquireFailure) beforeRoute(message string) {
	f.warnProviderCleanup(leaseapp.MarkFailedAcquireBeforeRouteCleanup(f.ctx, f.providerClient, f.session))
	f.save(message)
}

func (f *leaseAcquireFailure) afterRoute(route leaseapp.SessionRoute, message string) {
	f.warnProviderCleanup(leaseapp.MarkFailedAcquireAfterRouteCleanup(f.ctx, f.coordinator.deps.dataPlane, route, f.providerClient, f.session))
	f.save(message)
}

func (f *leaseAcquireFailure) warnProviderCleanup(err error) {
	if err == nil {
		return
	}
	providerID := ""
	if f.providerClient != nil {
		providerID = f.providerClient.Name()
	}
	f.coordinator.warn("provider session cleanup failed", "provider_id", providerID, "account_id", f.req.GetAccountId())
}

func (f *leaseAcquireFailure) save(message string) {
	f.coordinator.saveFailedAcquireLeaseFact(f.ctx, f.req, f.providerAccountID, f.session, f.egress, f.listener, f.plan, message)
}
