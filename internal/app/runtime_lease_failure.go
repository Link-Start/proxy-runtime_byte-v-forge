package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (c leaseCoordinator) saveFailedAcquireLeaseFact(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, session *proxyruntimev1.ProxySession, egress *proxyruntimev1.ProxyEndpoint, listener *proxyruntimev1.EgressListener, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan, message string) {
	if req == nil || strings.TrimSpace(req.GetAccountId()) == "" {
		return
	}
	leaseID, err := c.newLeaseID()
	if err != nil {
		return
	}
	lease := leaseapp.NewFailedAcquireFact(leaseapp.FailedAcquireFactInput{
		LeaseID:           leaseID,
		AccountID:         req.GetAccountId(),
		Purpose:           req.GetPurpose(),
		ProviderAccountID: providerAccountID,
		Session:           session,
		Egress:            egress,
		Listener:          listener,
		SelectionPlan:     plan,
		Message:           message,
		AcquiredAt:        c.now(),
	})
	if err := c.deps.store.SaveLeaseFact(ctx, lease); err != nil {
		c.warn("save failed proxy lease fact failed", "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
	}
}
