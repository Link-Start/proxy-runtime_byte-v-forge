package sqlite

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) SaveLeaseFact(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	fact, err := store.PrepareDynamicLeaseFactSave(lease)
	if err != nil {
		return err
	}
	now := sqliteTime(s.clock.Now().UTC())
	_, err = s.db.ExecContext(ctx, `
INSERT INTO proxy_gateway_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(lease_id) DO UPDATE SET account_id=excluded.account_id, purpose=excluded.purpose, provider_account_id=excluded.provider_account_id, status=excluded.status, lease_json=excluded.lease_json, acquired_at=excluded.acquired_at, expires_at=excluded.expires_at, updated_at=excluded.updated_at
`, fact.LeaseID, fact.AccountID, fact.Purpose, fact.ProviderAccountID, fact.Status.String(), fact.JSON, sqliteTimestamp(lease.GetAcquiredAt()), sqliteTimestamp(lease.GetExpiresAt()), now, now)
	return err
}
