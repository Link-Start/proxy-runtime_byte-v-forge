package postgres

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/jackc/pgx/v5"

	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func scanLeaseFact(row pgx.Row) (*proxygatewayv1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	return store.DecodeDynamicLeaseFactJSON(raw)
}

func scanLeaseFacts(rows pgx.Rows) ([]*proxygatewayv1.ProxyDynamicLease, error) {
	out := []*proxygatewayv1.ProxyDynamicLease{}
	for rows.Next() {
		lease, err := scanLeaseFact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lease)
	}
	return out, rows.Err()
}
