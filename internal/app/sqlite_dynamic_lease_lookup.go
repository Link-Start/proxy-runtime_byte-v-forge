package app

import (
	"context"
	"database/sql"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

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
