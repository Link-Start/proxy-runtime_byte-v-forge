package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) leaseListener(ctx context.Context, accountID string, leaseID string) (config.EgressListener, error) {
	_ = ctx
	leaseID = firstNonEmpty(leaseID, accountID)
	id := "lease-" + shortHash(leaseID)
	return config.EgressListener{
		ID:       id,
		Addr:     r.cfg.LocalAddr,
		Protocol: r.cfg.LocalProtocol,
		Route:    config.ListenerRouteProvider,
		Username: proxyRouteUsername(leaseID),
		Password: r.cfg.LocalPassword,
		Labels: map[string]string{
			"mode":       "dynamic_ip_session_lease",
			"account_id": accountID,
			"lease_id":   leaseID,
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
	username := runtimeSafeID(accountID)
	if username == "" {
		username = shortHash(accountID)
	}
	return "acct-" + username
}
