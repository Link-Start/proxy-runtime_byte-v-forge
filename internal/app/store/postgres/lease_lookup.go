package postgres

import (
	"context"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/jackc/pgx/v5"
)

func (s *Store) LeaseFactByID(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	row := s.pool.QueryRow(ctx, `SELECT lease_json::text FROM proxy_runtime_dynamic_leases WHERE lease_id=$1`, strings.TrimSpace(leaseID))
	return scanLeaseFact(row)
}

func (s *Store) ActiveLeaseFact(ctx context.Context, accountID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	row := s.pool.QueryRow(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE account_id=$1
	AND status=$2
	AND `+postgresLeaseActiveUntilNowPredicate+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC
LIMIT 1
`, strings.TrimSpace(accountID), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	return scanLeaseFact(row)
}

func (s *Store) ActiveLeaseFactBySession(ctx context.Context, accountID string, purpose string, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, pgx.ErrNoRows
	}
	args := []any{strings.TrimSpace(accountID), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sessionID}
	conditions := []string{`account_id=$1`, `status=$2`, postgresLeaseActiveUntilNowPredicate, `lease_json #>> '{session,sessionId}' = $3`}
	if trimmed := strings.TrimSpace(purpose); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("purpose=$%d", len(args)))
	}
	row := s.pool.QueryRow(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE `+strings.Join(conditions, " AND ")+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT 1
`, args...)
	return scanLeaseFact(row)
}

func (s *Store) ActiveLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, true)
}

func (s *Store) LatestLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, false)
}

func (s *Store) leaseFactByAccount(ctx context.Context, accountID string, purpose string, activeOnly bool) (*proxyruntimev1.ProxyDynamicLease, error) {
	args := []any{strings.TrimSpace(accountID)}
	conditions := []string{`account_id=$1`}
	if trimmed := strings.TrimSpace(purpose); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("purpose=$%d", len(args)))
	}
	if activeOnly {
		args = append(args, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
		conditions = append(conditions, fmt.Sprintf("status=$%d", len(args)), postgresLeaseActiveUntilNowPredicate)
	}
	row := s.pool.QueryRow(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE `+strings.Join(conditions, " AND ")+`
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC
LIMIT 1
`, args...)
	return scanLeaseFact(row)
}
