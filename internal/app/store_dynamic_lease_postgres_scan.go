package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/jackc/pgx/v5"

	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func scanLeaseFact(row pgx.Row) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	return store.DecodeDynamicLeaseFactJSON(raw)
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
