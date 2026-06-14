package lease

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func SaveFailedAcquireFact(ctx context.Context, store OrchestrationStore, ids IDGenerator, clock Clock, req *proxyruntimev1.AcquireProxyLeaseRequest, providerAccountID string, session *proxyruntimev1.ProxySession, egress *proxyruntimev1.ProxyEndpoint, listener *proxyruntimev1.EgressListener, plan *proxyruntimev1.ProxyDynamicIPSelectionPlan, message string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if req == nil || strings.TrimSpace(req.GetAccountId()) == "" {
		return nil, nil
	}
	if ids == nil {
		return nil, nil
	}
	if clock == nil {
		clock = SystemClock{}
	}
	leaseID, err := ids.NewLeaseID()
	if err != nil {
		return nil, nil
	}
	lease := NewFailedAcquireFact(FailedAcquireFactInput{
		LeaseID:           leaseID,
		AccountID:         req.GetAccountId(),
		Purpose:           req.GetPurpose(),
		ProviderAccountID: providerAccountID,
		Session:           session,
		Egress:            egress,
		Listener:          listener,
		SelectionPlan:     plan,
		Message:           message,
		AcquiredAt:        clock.Now(),
	})
	return lease, store.SaveLeaseFact(ctx, lease)
}
