package app

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (s *PostgresStore) ListActiveLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND (expires_at IS NULL OR expires_at > now())
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $2
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) ListRecentLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $1
`, leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) RecentLeaseFacts(ctx context.Context, since time.Time, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
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
