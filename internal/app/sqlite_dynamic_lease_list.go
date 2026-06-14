package app

import (
	"context"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

const sqliteCleanupPendingLeasePredicate = `(
  json_extract(lease_json, '$.session.labels.route_cleanup_pending')='true'
  OR json_extract(lease_json, '$.session.labels.provider_cleanup_pending')='true'
)`

func (s *SQLiteStore) ListActiveLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND (expires_at='' OR expires_at>?)
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()), leaseapp.NormalizeListLimit(limit))
}

func (s *SQLiteStore) ListRecentLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT ?
`, leaseapp.NormalizeListLimit(limit))
}

func (s *SQLiteStore) RecentLeaseFacts(ctx context.Context, since time.Time, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
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

func (s *SQLiteStore) ProviderAccountHasBlockingLease(ctx context.Context, providerAccountID string) (bool, error) {
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
      (status=? AND (expires_at='' OR expires_at>?))
      OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
    )
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *SQLiteStore) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	leases, err := s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=?
  AND (
    (status=? AND (expires_at='' OR expires_at>?))
    OR (status=? AND `+sqliteCleanupPendingLeasePredicate+`)
  )
ORDER BY acquired_at DESC, updated_at DESC, lease_id
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	out := filterLeaseFacts(leases, func(lease *proxyruntimev1.ProxyDynamicLease) bool {
		return strings.TrimSpace(lease.GetProviderAccountId()) == providerAccountID && (leaseapp.ActiveAt(lease, now) || leaseapp.CleanupPending(lease))
	})
	sortLeaseFactsByAcquiredDesc(out)
	return out, nil
}

func (s *SQLiteStore) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	leases, err := s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND `+sqliteCleanupPendingLeasePredicate+`
ORDER BY acquired_at ASC, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	return filterLeaseFacts(leases, leaseapp.CleanupPending), nil
}

func (s *SQLiteStore) ListRestorableLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND (expires_at='' OR expires_at>?)
ORDER BY acquired_at DESC, updated_at DESC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()))
}

func (s *SQLiteStore) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE status=? AND expires_at!='' AND expires_at<=?
ORDER BY expires_at ASC, acquired_at ASC, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()))
}
