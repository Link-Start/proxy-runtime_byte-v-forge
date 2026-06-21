package sqlite

import (
	"context"
	"database/sql"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (s *Store) LeaseFactByID(ctx context.Context, leaseID string) (*proxygatewayv1.ProxyDynamicLease, error) {
	row := s.db.QueryRowContext(ctx, `SELECT lease_json FROM proxy_gateway_dynamic_leases WHERE lease_id=?`, strings.TrimSpace(leaseID))
	return scanSQLiteLeaseFact(row)
}

func (s *Store) ActiveLeaseFact(ctx context.Context, accountID string) (*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, "", true)
}

func (s *Store) ActiveLeaseFactBySession(ctx context.Context, accountID string, purpose string, sessionID string) (*proxygatewayv1.ProxyDynamicLease, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sql.ErrNoRows
	}
	query := `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE account_id=? AND status=? AND ` + sqliteLeaseActiveUntilPredicate + ` AND json_extract(lease_json, '$.session.sessionId')=?`
	args := []any{strings.TrimSpace(accountID), proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()), sessionID}
	if purpose = strings.TrimSpace(purpose); purpose != "" {
		query += ` AND purpose=?`
		args = append(args, purpose)
	}
	query += ` ORDER BY acquired_at DESC, updated_at DESC, lease_id LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanSQLiteLeaseFact(row)
}

func (s *Store) ActiveLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, true)
}

func (s *Store) LatestLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxygatewayv1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, false)
}

func (s *Store) leaseFactByAccount(ctx context.Context, accountID string, purpose string, activeOnly bool) (*proxygatewayv1.ProxyDynamicLease, error) {
	accountID = strings.TrimSpace(accountID)
	purpose = strings.TrimSpace(purpose)
	query := `
SELECT lease_json
FROM proxy_gateway_dynamic_leases
WHERE account_id=?`
	args := []any{accountID}
	if purpose != "" {
		query += ` AND purpose=?`
		args = append(args, purpose)
	}
	if activeOnly {
		query += ` AND status=? AND ` + sqliteLeaseActiveUntilPredicate
		args = append(args, proxygatewayv1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sqliteTime(s.clock.Now().UTC()))
	}
	query += ` ORDER BY acquired_at DESC, updated_at DESC, lease_id LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanSQLiteLeaseFact(row)
}
