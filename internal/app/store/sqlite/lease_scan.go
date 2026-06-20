package sqlite

import (
	"context"
	"database/sql"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *Store) leaseFactsByQuery(ctx context.Context, query string, args ...any) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSQLiteLeaseFacts(rows)
}

func scanSQLiteLeaseFact(row interface{ Scan(...any) error }) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return store.DecodeDynamicLeaseFactJSON(raw)
}

func scanSQLiteLeaseFacts(rows *sql.Rows) ([]*proxyruntimev1.ProxyDynamicLease, error) {
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
