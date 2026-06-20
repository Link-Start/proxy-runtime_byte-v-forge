package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *PostgresStore) WriteSecret(ctx context.Context, req secretref.WriteRequest) (*commonv1.SecretRef, error) {
	if s == nil {
		return nil, errors.New("secret store is not configured")
	}
	if strings.TrimSpace(req.Value) == "" {
		return nil, errors.New("secret value is required")
	}
	provider := appcore.FirstNonEmpty(req.Provider, "proxy-runtime")
	purpose := strings.TrimSpace(req.Purpose)
	if purpose == "" {
		return nil, errors.New("secret purpose is required")
	}
	secretID := strings.TrimSpace(req.SecretID)
	if secretID == "" {
		generated, err := store.GeneratedSecretID(provider, purpose)
		if err != nil {
			return nil, err
		}
		secretID = generated
	}
	payload, err := s.box.Seal([]byte(req.Value))
	if err != nil {
		return nil, err
	}
	var expiresAt *time.Time
	if !req.ExpiresAt.IsZero() {
		expiresAt = &req.ExpiresAt
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO proxy_runtime_secrets (secret_id, provider, purpose, secret_payload, expires_at)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (secret_id) DO UPDATE SET provider=EXCLUDED.provider, purpose=EXCLUDED.purpose, secret_payload=EXCLUDED.secret_payload, expires_at=EXCLUDED.expires_at, updated_at=now()
`, secretID, provider, purpose, payload, expiresAt)
	if err != nil {
		return nil, err
	}
	return secretref.New(provider, purpose, secretID, req.ExpiresAt), nil
}

func (s *PostgresStore) ResolveSecret(ctx context.Context, ref *commonv1.SecretRef) (string, error) {
	if s == nil {
		return "", errors.New("secret store is not configured")
	}
	if err := secretref.Validate(ref); err != nil {
		return "", err
	}
	var provider, purpose, payload string
	var expiresAt pgtype.Timestamptz
	err := s.pool.QueryRow(ctx, `
SELECT provider, purpose, secret_payload, expires_at
FROM proxy_runtime_secrets
WHERE secret_id=$1
`, strings.TrimSpace(ref.GetSecretId())).Scan(&provider, &purpose, &payload, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("secret ref is not resolvable: %s", secretref.Display(ref))
	}
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ref.GetProvider()) != "" && strings.TrimSpace(ref.GetProvider()) != provider {
		return "", errors.New("secret provider mismatch")
	}
	if strings.TrimSpace(ref.GetPurpose()) != "" && strings.TrimSpace(ref.GetPurpose()) != purpose {
		return "", errors.New("secret purpose mismatch")
	}
	if expiresAt.Valid && s.clock.Now().After(expiresAt.Time) {
		return "", errors.New("secret ref is expired")
	}
	plain, err := s.box.Open(payload)
	if err != nil {
		return "", err
	}
	if len(plain) == 0 {
		return "", errors.New("secret payload is empty")
	}
	return string(plain), nil
}
