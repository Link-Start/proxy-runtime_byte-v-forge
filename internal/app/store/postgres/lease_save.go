package postgres

import (
	"context"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) SaveLeaseFact(ctx context.Context, lease *proxygatewayv1.ProxyDynamicLease) error {
	fact, err := store.PrepareDynamicLeaseFactSave(lease)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO proxy_gateway_dynamic_leases (lease_id, account_id, purpose, provider_account_id, status, lease_json, acquired_at, expires_at)
VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)
ON CONFLICT (lease_id) DO UPDATE SET account_id=EXCLUDED.account_id, purpose=EXCLUDED.purpose, provider_account_id=EXCLUDED.provider_account_id, status=EXCLUDED.status, lease_json=EXCLUDED.lease_json, acquired_at=EXCLUDED.acquired_at, expires_at=EXCLUDED.expires_at, updated_at=now()
`, fact.LeaseID, fact.AccountID, fact.Purpose, fact.ProviderAccountID, fact.Status.String(), fact.JSON, timestampValue(lease.GetAcquiredAt()), timestampValue(lease.GetExpiresAt()))
	return err
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
