package app

import (
	"context"
	"fmt"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (c leaseCoordinator) activeLeaseByRequest(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if c.deps.store == nil {
		return nil, fmt.Errorf("lease store is required")
	}
	if sessionID != "" {
		return c.deps.store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return c.deps.store.ActiveLeaseFactByAccount(ctx, req.GetAccountId(), req.GetPurpose())
}
