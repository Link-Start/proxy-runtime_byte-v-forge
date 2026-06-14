package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) newFailedAcquireRecorder(req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, providerClient leaseapp.SessionProvider, session *proxyruntimev1.ProxySession, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan) *leaseapp.FailedAcquireRecorder {
	return leaseapp.NewFailedAcquireRecorder(leaseapp.FailedAcquireRecorderInput{
		Store:             c.deps.store,
		IDs:               c.deps.ids,
		Clock:             c.deps.clock,
		DataPlane:         c.deps.dataPlane,
		Logger:            c.deps.logger,
		Request:           req,
		ProviderAccountID: providerAccountID,
		ProviderClient:    providerClient,
		Session:           session,
		SelectionPlan:     plan,
	})
}
