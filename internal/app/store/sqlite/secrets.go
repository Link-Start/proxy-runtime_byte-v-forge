package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/common/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) WriteSecret(ctx context.Context, req secretref.WriteRequest) (*commonv1.SecretRef, error) {
	if s == nil {
		return nil, errors.New("secret store is not configured")
	}
	if strings.TrimSpace(req.Value) == "" {
		return nil, errors.New("secret value is required")
	}
	provider := appcore.FirstNonEmpty(req.Provider, "proxy-gateway")
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
	now := sqliteTime(s.clock.Now().UTC())
	_, err = s.db.ExecContext(ctx, `
INSERT INTO proxy_gateway_secrets (secret_id, provider, purpose, secret_payload, expires_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(secret_id) DO UPDATE SET provider=excluded.provider, purpose=excluded.purpose, secret_payload=excluded.secret_payload, expires_at=excluded.expires_at, updated_at=excluded.updated_at
`, secretID, provider, purpose, payload, sqliteTime(req.ExpiresAt), now, now)
	if err != nil {
		return nil, err
	}
	return secretref.New(provider, purpose, secretID, req.ExpiresAt), nil
}

func (s *Store) ResolveSecret(ctx context.Context, ref *commonv1.SecretRef) (string, error) {
	if s == nil {
		return "", errors.New("secret store is not configured")
	}
	if err := secretref.Validate(ref); err != nil {
		return "", err
	}
	var provider, purpose, payload, expiresAtRaw string
	err := s.db.QueryRowContext(ctx, `SELECT provider, purpose, secret_payload, expires_at FROM proxy_gateway_secrets WHERE secret_id=?`, strings.TrimSpace(ref.GetSecretId())).Scan(&provider, &purpose, &payload, &expiresAtRaw)
	if errors.Is(err, sql.ErrNoRows) {
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
	if expiresAt := parseSQLiteTime(expiresAtRaw); !expiresAt.IsZero() && s.clock.Now().After(expiresAt) {
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
