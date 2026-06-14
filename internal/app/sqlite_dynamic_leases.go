package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const sqliteCleanupPendingLeasePredicate = `(
  json_extract(lease_json, '$.session.labels.route_cleanup_pending')='true'
  OR json_extract(lease_json, '$.session.labels.provider_cleanup_pending')='true'
)`

func (s *SQLiteStore) SaveLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	fact, err := prepareDynamicLeaseFactSave(lease)
	if err != nil {
		return err
	}
	now := sqliteTime(time.Now().UTC())
	_, err = s.db.ExecContext(ctx, `
INSERT INTO proxy_runtime_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(lease_id) DO UPDATE SET account_id=excluded.account_id, purpose=excluded.purpose, provider_account_id=excluded.provider_account_id, status=excluded.status, lease_json=excluded.lease_json, acquired_at=excluded.acquired_at, expires_at=excluded.expires_at, updated_at=excluded.updated_at
`, fact.LeaseID, fact.AccountID, fact.Purpose, fact.ProviderAccountID, fact.Status.String(), fact.JSON, sqliteTimestamp(lease.GetAcquiredAt()), sqliteTimestamp(lease.GetExpiresAt()), now, now)
	return err
}

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

func (s *SQLiteStore) LeaseFactByID(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	row := s.db.QueryRowContext(ctx, `SELECT lease_json FROM proxy_runtime_dynamic_leases WHERE lease_id=?`, strings.TrimSpace(leaseID))
	return scanSQLiteLeaseFact(row)
}

func (s *SQLiteStore) ActiveLeaseFact(ctx context.Context, accountID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, "", true)
}

func (s *SQLiteStore) ActiveLeaseFactBySession(ctx context.Context, accountID string, purpose string, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	query := `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE account_id=? AND status=? AND (expires_at='' OR expires_at>?) AND json_extract(lease_json, '$.session.sessionId')=?`
	args := []any{strings.TrimSpace(accountID), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()), sessionID}
	if purpose = strings.TrimSpace(purpose); purpose != "" {
		query += ` AND purpose=?`
		args = append(args, purpose)
	}
	query += ` ORDER BY acquired_at DESC, updated_at DESC, lease_id LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanSQLiteLeaseFact(row)
}

func (s *SQLiteStore) ActiveLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, true)
}

func (s *SQLiteStore) LatestLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, false)
}

func (s *SQLiteStore) leaseFactByAccount(ctx context.Context, accountID string, purpose string, activeOnly bool) (*proxyruntimev1.ProxyDynamicLease, error) {
	accountID = strings.TrimSpace(accountID)
	purpose = strings.TrimSpace(purpose)
	query := `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE account_id=?`
	args := []any{accountID}
	if purpose != "" {
		query += ` AND purpose=?`
		args = append(args, purpose)
	}
	if activeOnly {
		query += ` AND status=? AND (expires_at='' OR expires_at>?)`
		args = append(args, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC()))
	}
	query += ` ORDER BY acquired_at DESC, updated_at DESC, lease_id LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanSQLiteLeaseFact(row)
}

func scanSQLiteLeaseFact(row interface{ Scan(...any) error }) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return decodeDynamicLeaseFactJSON(raw)
}
