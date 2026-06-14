package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type leaseAcquireFailure struct {
	coordinator       leaseCoordinator
	ctx               context.Context
	req               *proxyruntimev1.AcquireProxyLeaseRequest
	providerAccountID string
	providerClient    provider.SessionProvider
	session           *proxyruntimev1.ProxySession
	listener          *proxyruntimev1.EgressListener
	egress            *proxyruntimev1.ProxyEndpoint
	plan              *proxyruntimev1.ProxyDynamicIPSelectionPlan
}

func newLeaseAcquireFailure(coordinator leaseCoordinator, ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, providerClient provider.SessionProvider, session *proxyruntimev1.ProxySession, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan) *leaseAcquireFailure {
	return &leaseAcquireFailure{coordinator: coordinator, ctx: ctx, req: req, providerAccountID: providerAccountID, providerClient: providerClient, session: session, plan: plan}
}

func (f *leaseAcquireFailure) beforeRoute(message string) {
	if f.cleanupProviderSession() {
		f.markCleanupPending(false, true)
	}
	f.save(message)
}

func (f *leaseAcquireFailure) afterRoute(route dataplane.SessionRoute, message string) {
	routeCleanupPending := f.coordinator.deps.dataPlane.DeleteSessionRoute(f.ctx, route) != nil
	providerCleanupPending := f.cleanupProviderSession()
	f.markCleanupPending(routeCleanupPending, providerCleanupPending)
	f.save(message)
}

func (f *leaseAcquireFailure) cleanupProviderSession() bool {
	if err := releaseProviderSession(f.ctx, f.providerClient, f.session); err != nil {
		f.coordinator.warn("provider session cleanup failed", "provider_id", f.providerClient.Name(), "account_id", f.req.GetAccountId())
		return true
	}
	return false
}

func (f *leaseAcquireFailure) markCleanupPending(routePending bool, providerPending bool) {
	lease := &proxyruntimev1.ProxyDynamicLease{Session: f.session}
	leaseapp.MarkCleanupPending(lease, routePending, providerPending, leaseapp.CleanupFinalFailed)
}

func (f *leaseAcquireFailure) save(message string) {
	f.coordinator.saveFailedAcquireLeaseFact(f.ctx, f.req, f.providerAccountID, f.session, f.egress, f.listener, f.plan, message)
}
