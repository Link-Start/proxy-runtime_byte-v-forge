package sqlite

import (
	"context"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
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
  FROM proxy_gateway_dynamic_leases
  WHERE provider_account_id=?
    AND (
      (status=? AND `+sqliteLeaseActiveUntilPredicate+`)
      OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
    )
)
`, providerAccountID, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *Store) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string, limit int) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE provider_account_id=?
  AND (
    (status=? AND `+sqliteLeaseActiveUntilPredicate+`)
    OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
  )
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, providerAccountID, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String(), kernel.NormalizeBlockingLeaseFactLimit(limit))
}
