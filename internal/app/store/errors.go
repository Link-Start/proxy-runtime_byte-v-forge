package store

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

// IsNotFound reports whether err represents a missing row from either the
// SQLite or PostgreSQL backend.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}
