package sqlite

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (s *Store) ListActiveLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND `+sqliteLeaseActiveUntilPredicate+`
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), leaseapp.NormalizeListLimit(limit))
}

func (s *Store) ListRecentLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, leaseapp.NormalizeListLimit(limit))
}

func (s *Store) ListHistoryLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status!=? OR `+sqliteLeaseExpiredByPredicate+`
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), leaseapp.NormalizeListLimit(limit))
}

func (s *Store) RecentLeaseFacts(ctx context.Context, since time.Time, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE acquired_at!='' AND acquired_at>=?
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, sqliteTime(since.UTC()), limit)
}
