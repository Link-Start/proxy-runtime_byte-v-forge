package app

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

func isStoreNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}
