package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func ActiveLeaseByRequest(ctx context.Context, store OrchestrationStore, req *proxyruntimev1.AcquireProxyLeaseRequest, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	if store == nil {
		return nil, errors.New("lease store is required")
	}
	if sessionID != "" {
		return store.ActiveLeaseFactBySession(ctx, req.GetAccountId(), req.GetPurpose(), sessionID)
	}
	return store.ActiveLeaseFactByAccount(ctx, req.GetAccountId(), req.GetPurpose())
}
