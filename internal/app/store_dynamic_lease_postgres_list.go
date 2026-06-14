package app

import (
	"context"
	"strings"
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

func (s *PostgresStore) ProviderAccountHasBlockingLease(ctx context.Context, providerAccountID string) (bool, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return false, nil
	}
	var exists bool
	err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM proxy_runtime_dynamic_leases
	WHERE provider_account_id=$1
		AND (
			(status=$2 AND (expires_at IS NULL OR expires_at > now()))
			OR (
				status=$3
				AND (
					lease_json #>> '{session,labels,route_cleanup_pending}' = 'true'
					OR lease_json #>> '{session,labels,provider_cleanup_pending}' = 'true'
				)
			)
		)
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *PostgresStore) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=$1
	AND (
		(status=$2 AND (expires_at IS NULL OR expires_at > now()))
		OR (
			status=$3
			AND (
				lease_json #>> '{session,labels,route_cleanup_pending}' = 'true'
				OR lease_json #>> '{session,labels,provider_cleanup_pending}' = 'true'
			)
		)
	)
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND (
		lease_json #>> '{session,labels,route_cleanup_pending}' = 'true'
		OR lease_json #>> '{session,labels,provider_cleanup_pending}' = 'true'
	)
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
	AND (expires_at IS NULL OR expires_at > now())
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
	AND expires_at IS NOT NULL
	AND expires_at <= now()
ORDER BY expires_at ASC, acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}
