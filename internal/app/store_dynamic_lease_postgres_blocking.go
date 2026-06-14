package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
