package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (s *SQLiteStore) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND `+sqliteCleanupPendingLeasePredicate+`
ORDER BY acquired_at ASC, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
}

func (s *SQLiteStore) ListRestorableLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND `+sqliteLeaseActiveUntilPredicate+`
ORDER BY acquired_at DESC, updated_at DESC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()))
}

func (s *SQLiteStore) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND `+sqliteLeaseExpiredByPredicate+`
ORDER BY expires_at ASC, acquired_at ASC, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()))
}
