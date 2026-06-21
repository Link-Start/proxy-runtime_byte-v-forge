package sqlite

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (s *Store) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE status=? AND `+sqliteCleanupPendingLeasePredicate+`
ORDER BY acquired_at ASC, updated_at ASC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
}

func (s *Store) ListRestorableLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE status=? AND `+sqliteLeaseActiveUntilPredicate+`
ORDER BY acquired_at DESC, updated_at DESC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()))
}

func (s *Store) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE status=? AND `+sqliteLeaseExpiredByPredicate+`
ORDER BY expires_at ASC, acquired_at ASC, updated_at ASC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()))
}
