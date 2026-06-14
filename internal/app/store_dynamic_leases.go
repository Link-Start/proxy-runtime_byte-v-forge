package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PostgresStore) SaveLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	fact, err := prepareDynamicLeaseFactSave(lease)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO proxy_runtime_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at)
VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)
ON CONFLICT (lease_id) DO UPDATE SET account_id=EXCLUDED.account_id, purpose=EXCLUDED.purpose, provider_account_id=EXCLUDED.provider_account_id, status=EXCLUDED.status, lease_json=EXCLUDED.lease_json, acquired_at=EXCLUDED.acquired_at, expires_at=EXCLUDED.expires_at, updated_at=now()
`, fact.LeaseID, fact.AccountID, fact.Purpose, fact.ProviderAccountID, fact.Status.String(), fact.JSON, timestampValue(lease.GetAcquiredAt()), timestampValue(lease.GetExpiresAt()))
	return err
}

func (s *PostgresStore) ListActiveLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND (expires_at IS NULL OR expires_at > now())
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $2
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) ListRecentLeaseFacts(ctx context.Context, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
LIMIT $1
`, leaseapp.NormalizeListLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaseFacts(rows)
}

func (s *PostgresStore) RecentLeaseFacts(ctx context.Context, since time.Time, limit int) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE acquired_at IS NOT NULL
	AND acquired_at >= $1
ORDER BY acquired_at DESC, updated_at DESC, lease_id
LIMIT $2
`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ProviderAccountHasBlockingLease(ctx context.Context, providerAccountID string) (bool, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return false, nil
	}
	var exists bool
	err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM proxy_runtime_dynamic_leases
	WHERE provider_account_id=$1
		AND (
			(status=$2 AND (expires_at IS NULL OR expires_at > now()))
			OR (
				status=$3
				AND (
					lease_json #>> '{session,labels,route_cleanup_pending}' = 'true'
					OR lease_json #>> '{session,labels,provider_cleanup_pending}' = 'true'
				)
			)
		)
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String()).Scan(&exists)
	return exists, err
}

func (s *PostgresStore) BlockingLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=$1
	AND (
		(status=$2 AND (expires_at IS NULL OR expires_at > now()))
		OR status=$3
	)
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		if leaseapp.ActiveAt(lease, time.Now().UTC()) || leaseapp.CleanupPending(lease) {
			out = append(out, lease)
		}
	}
	return out, rows.Err()
}

func (s *PostgresStore) CleanupPendingLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
ORDER BY acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_FAILED.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		if leaseapp.CleanupPending(lease) {
			out = append(out, lease)
		}
	}
	return out, rows.Err()
}

func (s *PostgresStore) ListRestorableLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND (expires_at IS NULL OR expires_at > now())
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ExpiredActiveLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND expires_at IS NOT NULL
	AND expires_at <= now()
ORDER BY expires_at ASC, acquired_at ASC NULLS LAST, updated_at ASC, lease_id
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}

func (s *PostgresStore) LeaseFactByID(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	row := s.pool.QueryRow(ctx, `SELECT lease_json::text FROM proxy_runtime_dynamic_leases WHERE lease_id=$1`, strings.TrimSpace(leaseID))
	return scanLeaseFact(row)
}

func (s *PostgresStore) ActiveLeaseFact(ctx context.Context, accountID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	row := s.pool.QueryRow(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE account_id=$1
	AND status=$2
	AND (expires_at IS NULL OR expires_at > now())
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC
LIMIT 1
`, strings.TrimSpace(accountID), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	return scanLeaseFact(row)
}

func (s *PostgresStore) ActiveLeaseFactBySession(ctx context.Context, accountID string, purpose string, sessionID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, pgx.ErrNoRows
	}
	args := []any{strings.TrimSpace(accountID), proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String(), sessionID}
	conditions := []string{`account_id=$1`, `status=$2`, `(expires_at IS NULL OR expires_at > now())`, `lease_json #>> '{session,sessionId}' = $3`}
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

func (s *PostgresStore) ActiveLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, true)
}

func (s *PostgresStore) LatestLeaseFactByAccount(ctx context.Context, accountID string, purpose string) (*proxyruntimev1.ProxyDynamicLease, error) {
	return s.leaseFactByAccount(ctx, accountID, purpose, false)
}

func (s *PostgresStore) leaseFactByAccount(ctx context.Context, accountID string, purpose string, activeOnly bool) (*proxyruntimev1.ProxyDynamicLease, error) {
	args := []any{strings.TrimSpace(accountID)}
	conditions := []string{`account_id=$1`}
	if trimmed := strings.TrimSpace(purpose); trimmed != "" {
		args = append(args, trimmed)
		conditions = append(conditions, fmt.Sprintf("purpose=$%d", len(args)))
	}
	if activeOnly {
		args = append(args, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
		conditions = append(conditions, fmt.Sprintf("status=$%d", len(args)), `(expires_at IS NULL OR expires_at > now())`)
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

func scanLeaseFact(row pgx.Row) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
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

func scanLeaseFacts(rows pgx.Rows) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	out := []*proxyruntimev1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}

func timestampValue(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	t := value.AsTime()
	if t.IsZero() {
		return nil
	}
	return &t
}
