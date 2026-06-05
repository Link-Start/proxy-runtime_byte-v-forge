package app

import (
	"context"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/randx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c leaseCoordinator) saveFailedAcquireLeaseFact(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, session *proxyruntimev1.ProxySession, egress *proxyruntimev1.ProxyEndpoint, listener *proxyruntimev1.EgressListener, plan *proxyruntimev1.EgressRoutePlan, message string) {
	r := c.runtime
	if req == nil || strings.TrimSpace(req.GetAccountId()) == "" {
		return
	}
	leaseID, err := randx.Hex(12)
	if err != nil {
		return
	}
	lease := &proxyruntimev1.ProxyDynamicLease{
		LeaseId:           leaseID,
		AccountId:         strings.TrimSpace(req.GetAccountId()),
		Purpose:           firstNonEmpty(req.GetPurpose(), "general"),
		ProviderAccountId: strings.TrimSpace(providerAccountID),
		Status:            proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED,
		Session:           session,
		Egress:            egress,
		Listener:          listener,
		AcquiredAt:        timestamppb.New(time.Now().UTC()),
		RoutePlan:         plan,
		ErrorMessage:      strings.TrimSpace(message),
	}
	if session != nil {
		lease.ExpiresAt = session.GetExpiresAt()
	}
	if lease.ErrorMessage == "" {
		lease.ErrorMessage = "lease acquire failed"
	}
	if err := r.store.SaveLeaseFact(ctx, lease); err != nil {
		r.logger.Warn("save failed proxy lease fact failed", "account_id", lease.GetAccountId(), "provider_account_id", lease.GetProviderAccountId())
	}
}
