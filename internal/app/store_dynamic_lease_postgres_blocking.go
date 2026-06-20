package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
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
			(status=$2 AND `+postgresLeaseActiveUntilNowPredicate+`)
			OR (status=$3 AND `+postgresLeaseCleanupPendingPredicate+`)
		)
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *PostgresStore) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=$1
	AND (
		(status=$2 AND `+postgresLeaseActiveUntilNowPredicate+`)
		OR (status=$3 AND `+postgresLeaseCleanupPendingPredicate+`)
	)
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $4
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String(), settingscore.NormalizeBlockingLeaseFactLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}
