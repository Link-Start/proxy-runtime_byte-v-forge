package app

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *SQLiteStore) SaveLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil || strings.TrimSpace(lease.GetLeaseId()) == "" {
		return errors.New("lease_id is required")
	}
	if strings.TrimSpace(lease.GetAccountId()) == "" {
		return errors.New("lease account_id is required")
	}
	status := lease.GetStatus()
	if status == proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_UNSPECIFIED {
		status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE
		lease.Status = status
	}
	data, err := protojsoncodec.Marshal(lease)
	if err != nil {
		return err
	}
	now := sqliteTime(time.Now().UTC())
	_, err = s.db.ExecContext(ctx, `
INSERT INTO proxy_runtime_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(lease_id) DO UPDATE SET account_id=excluded.account_id, purpose=excluded.purpose, provider_account_id=excluded.provider_account_id, status=excluded.status, lease_json=excluded.lease_json, acquired_at=excluded.acquired_at, expires_at=excluded.expires_at, updated_at=excluded.updated_at
`, strings.TrimSpace(lease.GetLeaseId()), strings.TrimSpace(lease.GetAccountId()), strings.TrimSpace(lease.GetPurpose()), strings.TrimSpace(lease.GetProviderAccountId()), status.String(), string(data), sqliteTimestamp(lease.GetAcquiredAt()), sqliteTimestamp(lease.GetExpiresAt()), now, now)
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
	leases, err := s.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID)
	if err != nil {
		return false, err
	}
	return len(leases) > 0, nil
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
    OR status=?
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
WHERE status=?
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
	leases, err := s.activeLeaseFactsByAccount(ctx, accountID, purpose)
	if err != nil {
		return nil, err
	}
	for _, lease := range leases {
		if strings.TrimSpace(lease.GetSession().GetSessionId()) == sessionID {
			return lease, nil
		}
	}
	return nil, sql.ErrNoRows
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

func (s *SQLiteStore) activeLeaseFactsByAccount(ctx context.Context, accountID string, purpose string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	accountID = strings.TrimSpace(accountID)
	purpose = strings.TrimSpace(purpose)
	query := `
SELECT lease_json
FROM proxy_runtime_dynamic_leases
WHERE account_id=? AND status=? AND (expires_at='' OR expires_at>?)`
	args := []any{accountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(time.Now().UTC())}
	if purpose != "" {
		query += ` AND purpose=?`
		args = append(args, purpose)
	}
	query += ` ORDER BY acquired_at DESC, updated_at DESC, lease_id`
	return s.leaseFactsByQuery(ctx, query, args...)
}

func (s *SQLiteStore) allLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactsByQuery(ctx, `SELECT lease_json FROM proxy_runtime_dynamic_leases ORDER BY updated_at DESC, lease_id`)
}

func (s *SQLiteStore) leaseFactsByQuery(ctx context.Context, query string, args ...any) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanSQLiteLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}

func scanSQLiteLeaseFact(row interface{ Scan(...any) error }) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	lease := &proxyruntimev1.ProxyDynamicLease{}
	if strings.TrimSpace(raw) == "" {
		return lease, nil
	}
	if err := protojsoncodec.Unmarshal([]byte(raw), lease); err != nil {
		return nil, err
	}
	return lease, nil
}

func filterLeaseFacts(in []*proxyruntimev1.ProxyDynamicLease, keep func(*proxyruntimev1.ProxyDynamicLease) bool) []*proxyruntimev1.ProxyDynamicLease {
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for _, lease := range in {
		if keep(lease) {
			out = append(out, lease)
		}
	}
	return out
}

func sortLeaseFactsByAcquiredDesc(leases []*proxyruntimev1.ProxyDynamicLease) {
	sort.SliceStable(leases, func(left, right int) bool {
		return leaseSortTime(leases[left]).After(leaseSortTime(leases[right]))
	})
}

func leaseSortTime(lease *proxyruntimev1.ProxyDynamicLease) time.Time {
	if lease.GetAcquiredAt() != nil {
		return lease.GetAcquiredAt().AsTime()
	}
	if lease.GetExpiresAt() != nil {
		return lease.GetExpiresAt().AsTime()
	}
	return time.Time{}
}

func sqliteTimestamp(value *timestamppb.Timestamp) string {
	if value == nil {
		return ""
	}
	return sqliteTime(value.AsTime())
}
