package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/protojsonx"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PostgresStore) SaveLeaseFact(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
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
	data, err := protojsonx.Marshal(lease)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO proxy_runtime_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at)
VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)
ON CONFLICT (lease_id) DO UPDATE SET account_id=EXCLUDED.account_id, purpose=EXCLUDED.purpose, provider_account_id=EXCLUDED.provider_account_id, status=EXCLUDED.status, lease_json=EXCLUDED.lease_json, acquired_at=EXCLUDED.acquired_at, expires_at=EXCLUDED.expires_at, updated_at=now()
`, strings.TrimSpace(lease.GetLeaseId()), strings.TrimSpace(lease.GetAccountId()), strings.TrimSpace(lease.GetPurpose()), strings.TrimSpace(lease.GetProviderAccountId()), status.String(), string(data), timestampValue(lease.GetAcquiredAt()), timestampValue(lease.GetExpiresAt()))
	return err
}

func (s *PostgresStore) ListLeaseFacts(ctx context.Context, includeInactive bool) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	query := `SELECT lease_json::text FROM proxy_runtime_dynamic_leases`
	args := []any{}
	if !includeInactive {
		query += ` WHERE status=$1 AND (expires_at IS NULL OR expires_at > now())`
		args = append(args, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	}
	query += ` ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id`
	rows, err := s.pool.Query(ctx, query, args...)
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

func (s *PostgresStore) ActiveProviderAccountIDs(ctx context.Context) (map[string]struct{}, error) {
	rows, err := s.pool.Query(ctx, `
SELECT DISTINCT provider_account_id
FROM proxy_runtime_dynamic_leases
WHERE status=$1
	AND provider_account_id <> ''
	AND (expires_at IS NULL OR expires_at > now())
`, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]struct{}{}
	for rows.Next() {
		var providerAccountID string
		if err := rows.Scan(&providerAccountID); err != nil {
			return nil, err
		}
		if providerAccountID = strings.TrimSpace(providerAccountID); providerAccountID != "" {
			out[providerAccountID] = struct{}{}
		}
	}
	return out, rows.Err()
}

func (s *PostgresStore) ProviderAccountHasActiveLease(ctx context.Context, providerAccountID string) (bool, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return false, nil
	}
	var active bool
	err := s.pool.QueryRow(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM proxy_runtime_dynamic_leases
	WHERE provider_account_id=$1
		AND status=$2
		AND (expires_at IS NULL OR expires_at > now())
)
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String()).Scan(&active)
	return active, err
}

func (s *PostgresStore) ProviderAccountHasBlockingLease(ctx context.Context, providerAccountID string) (bool, error) {
	leases, err := s.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID)
	if err != nil {
		return false, err
	}
	return len(leases) > 0, nil
}

func (s *PostgresStore) ActiveLeaseFactsByProviderAccount(ctx context.Context, providerAccountID string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE provider_account_id=$1
	AND status=$2
	AND (expires_at IS NULL OR expires_at > now())
ORDER BY acquired_at DESC NULLS LAST, updated_at DESC, lease_id
`, providerAccountID, proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE.String())
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
		if leaseActive(lease, time.Now().UTC()) || leaseCleanupPending(lease) {
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
		if leaseCleanupPending(lease) {
			out = append(out, lease)
		}
	}
	return out, rows.Err()
}

func (s *PostgresStore) BlockingLeaseFactsBySource(ctx context.Context, sourceID string) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	_ = ctx
	_ = s
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return nil, nil
	}
	return nil, nil
}

func (s *PostgresStore) ListRestorableLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.pool.Query(ctx, `
SELECT lease_json::text
FROM proxy_runtime_dynamic_leases
WHERE status=$1
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
	if err := protojsonx.Unmarshal([]byte(raw), lease); err != nil {
		return nil, err
	}
	return lease, nil
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
