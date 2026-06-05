package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, accountID string) (config.EgressListener, error) {
	_ = ctx
	id := "lease-" + shortHash(accountID)
	return config.EgressListener{
		ID:       id,
		Addr:     r.cfg.LocalAddr,
		Protocol: r.cfg.LocalProtocol,
		Route:    config.ListenerRouteProvider,
		Username: proxyRouteUsername(accountID),
		Password: r.cfg.LocalPassword,
		Labels: map[string]string{
			"mode":       "dynamic_ip_session_lease",
			"account_id": accountID,
		},
	}, nil
}

func (r *Runtime) listenerReservedLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	leases, err := r.store.ListLeaseFacts(ctx, true)
	if err != nil {
		return nil, err
	}
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(leases))
	for _, lease := range leases {
		if lease.GetListener() == nil {
			continue
		}
		if lease.GetStatus() == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE || leaseRouteCleanupPending(lease) {
			out = append(out, lease)
		}
	}
	return out, nil
}

func proxyRouteUsername(accountID string) string {
	username := sourceSafeID(accountID)
	if username == "" {
		username = shortHash(accountID)
	}
	return "acct-" + username
}
