package app

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

type dynamicProfileSessionState struct {
	generation int64
	released   bool
}

func (s *PostgresStore) DynamicProfileSessionState(ctx context.Context, profileKey string) (dynamicProfileSessionState, error) {
	profileKey = strings.TrimSpace(profileKey)
	if profileKey == "" {
		return dynamicProfileSessionState{}, nil
	}
	var state dynamicProfileSessionState
	err := s.pool.QueryRow(ctx, `SELECT generation, released FROM proxy_runtime_dynamic_profile_sessions WHERE profile_key=$1`, profileKey).Scan(&state.generation, &state.released)
	if errors.Is(err, pgx.ErrNoRows) {
		return dynamicProfileSessionState{}, nil
	}
	return state, err
}

func (s *PostgresStore) MarkDynamicProfileSessionReleased(ctx context.Context, profileKey string) error {
	profileKey = strings.TrimSpace(profileKey)
	if profileKey == "" {
		return errors.New("dynamic profile session key is required")
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO proxy_runtime_dynamic_profile_sessions (profile_key, released)
VALUES ($1, true)
ON CONFLICT (profile_key) DO UPDATE
SET released=true, updated_at=now()
`, profileKey)
	return err
}
