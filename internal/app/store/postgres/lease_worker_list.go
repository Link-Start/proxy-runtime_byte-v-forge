package postgres

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (s *Store) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseCleanupPendingPredicate+`
ORDER BY acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *Store) ListRestorableLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseActiveUntilNowPredicate+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *Store) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseExpiredByNowPredicate+`
ORDER BY expires_at ASC, acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}
