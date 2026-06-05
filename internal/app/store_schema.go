package app

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

func (s *PostgresStore) applySchema(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply proxy-runtime schema: %w", err)
	}
	return nil
}
