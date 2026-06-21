package postgres

import (
	"context"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
)

func (s *Store) ListActiveLeaseFacts(ctx context.Context, limit int) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseActiveUntilNowPredicate+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $2
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *Store) ListRecentLeaseFacts(ctx context.Context, limit int) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $1
`, leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *Store) ListHistoryLeaseFacts(ctx context.Context, limit int) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE status<>$1
	OR `+postgresLeaseExpiredByNowPredicate+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $2
`, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *Store) RecentLeaseFacts(ctx context.Context, since time.Time, limit int) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_gateway_dynamic_leases
WHERE acquired_at IS NOT NULL
	AND acquired_at >= $1
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT $2
`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}
