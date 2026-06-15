package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *PostgresStore) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseCleanupPendingPredicate+`
ORDER BY acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) ListRestorableLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseActiveUntilNowPredicate+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND `+postgresLeaseExpiredByNowPredicate+`
ORDER BY expires_at ASC, acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}
