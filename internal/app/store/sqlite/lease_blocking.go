package sqlite

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func (s *Store) ProviderAccountHasBlockingLease(ctx context.Context, providerAccountID string) (bool, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return false, nil
	}
	var exists bool
	err := s.db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM proxy_runtime_dynamic_leases
  WHERE provider_account_id=?
    AND (
      (status=? AND `+sqliteLeaseActiveUntilPredicate+`)
      OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
    )
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *Store) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=?
  AND (
    (status=? AND `+sqliteLeaseActiveUntilPredicate+`)
    OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
  )
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String(), settingscore.NormalizeBlockingLeaseFactLimit(limit))
}
