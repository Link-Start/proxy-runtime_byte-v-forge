package app

import (
	"database/sql"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func scanSQLiteLeaseFact(row interface{ Scan(...any) error }) (*proxyruntimev1.ProxyDynamicLease, error) {
	var raw string
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return decodeDynamicLeaseFactJSON(raw)
}
