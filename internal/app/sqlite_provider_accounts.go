package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func (s *SQLiteStore) ListProviderAccounts(ctx context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at FROM proxy_runtime_provider_accounts ORDER BY account_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyProviderAccount{}
	for rows.Next() {
		record, err := scanSQLiteProviderAccount(rows)
		if err != nil {
			return nil, err
		}
		account, err := providerAccountToProto(ctx, s, s.box, record)
		if err != nil {
			return nil, err
		}
		out = append(out, account)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) UpsertProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error) {
	accountID := normalizeID(req.GetAccountId())
	if accountID == "" {
		generated, err := generatedID("dynacct")
		if err != nil {
			return nil, err
		}
		accountID = generated
	}
	existing, err := s.providerAccountRecord(ctx, accountID)
	if err != nil && !isStoreNotFound(err) {
		return nil, err
	}
	providerID := firstNonEmpty(strings.TrimSpace(req.GetProviderId()), existingProviderID(existing), accountproxy.ProviderTen24)
	if !s.accountProviders.IsSupported(providerID) {
		return nil, fmt.Errorf("unsupported provider_id %q", providerID)
	}
	dynamicProviderID := firstNonEmpty(runtimeSafeID(req.GetDynamicProviderId()), existingDynamicProviderID(existing))
	secret := existingCredentialSecret(existing)
	credential := providerCredential{}
	if current := credentialFromSecret(s.box, secret); current != nil {
		credential = *current
	}
	if req.GetClearPassword() {
		credential.PasswordSecretRef = nil
	}
	if username := strings.TrimSpace(req.GetUsername()); username != "" {
		credential.Username = username
	}
	if ref := cloneSecretRef(req.GetPasswordSecretRef(), "proxy-runtime", "dynamic_ip_provider_password"); ref != nil {
		credential.PasswordSecretRef = ref
	}
	if rawPassword := strings.TrimSpace(req.GetPasswordValue()); rawPassword != "" {
		ref, err := s.WriteSecret(ctx, secretref.WriteRequest{SecretID: secretref.StableID("proxy-runtime-provider-account-password", accountID), Provider: "proxy-runtime", Purpose: "dynamic_ip_provider_password", Value: rawPassword})
		if err != nil {
			return nil, err
		}
		credential.PasswordSecretRef = ref
	}
	if credential.Username != "" || credential.PasswordSecretRef != nil {
		payload, err := json.Marshal(credential)
		if err != nil {
			return nil, err
		}
		secret, err = s.box.Seal(payload)
		if err != nil {
			return nil, err
		}
	} else {
		secret = ""
	}
	enabled := req.GetEnabled()
	displayName := firstNonEmpty(req.GetDisplayName(), accountID)
	if enabled {
		cfg, err := providerConfigFromCredentialSecret(ctx, s, s.box, providerID, secret)
		if err != nil {
			return nil, fmt.Errorf("enabled provider account invalid: %w", err)
		}
		if err := s.accountProviders.Validate(cfg); err != nil {
			return nil, fmt.Errorf("enabled provider account invalid: %w", err)
		}
	}
	now := time.Now().UTC()
	createdAt := now
	if existing != nil && !existing.CreatedAt.IsZero() {
		createdAt = existing.CreatedAt
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO proxy_runtime_provider_accounts (account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?)
ON CONFLICT(account_id) DO UPDATE SET provider_id=excluded.provider_id, dynamic_provider_id=excluded.dynamic_provider_id, display_name=excluded.display_name, enabled=excluded.enabled, credential_secret=excluded.credential_secret, updated_at=excluded.updated_at
`, accountID, providerID, dynamicProviderID, displayName, sqliteBool(enabled), secret, sqliteTime(createdAt), sqliteTime(now))
	if err != nil {
		return nil, err
	}
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return providerAccountToProto(ctx, s, s.box, record)
}

func (s *SQLiteStore) DeleteProviderAccount(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM proxy_runtime_provider_accounts WHERE account_id=?`, normalizeID(accountID))
	return err
}

func (s *SQLiteStore) ProviderAccount(ctx context.Context, accountID string) (*proxyruntimev1.ProxyProviderAccount, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return providerAccountToProto(ctx, s, s.box, record)
}

func (s *SQLiteStore) ProviderAccountMutationState(ctx context.Context, accountID string) (providerAccountMutationState, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return providerAccountMutationState{}, err
	}
	credential := credentialFromSecret(s.box, record.CredentialSecret)
	state := providerAccountMutationState{ProviderID: record.ProviderID, DynamicProviderID: record.DynamicProviderID, PasswordConfigured: record.CredentialSecret != ""}
	if credential != nil {
		state.Username = credential.Username
		state.PasswordSecretRef = cloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password")
	}
	return state, nil
}

func (s *SQLiteStore) ProviderConfig(ctx context.Context, accountID string) (accountproxy.Config, string, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return accountproxy.Config{}, "", err
	}
	if !record.Enabled {
		return accountproxy.Config{}, "", errors.New("provider account is disabled")
	}
	cfg, err := providerConfigFromCredentialSecret(ctx, s, s.box, record.ProviderID, record.CredentialSecret)
	if err != nil {
		return accountproxy.Config{}, "", err
	}
	return cfg, record.AccountID, s.accountProviders.Validate(cfg)
}

func (s *SQLiteStore) DefaultProviderAccountID(ctx context.Context) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT account_id FROM proxy_runtime_provider_accounts WHERE enabled=1 ORDER BY updated_at DESC, account_id LIMIT 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	return id, err
}

func (s *SQLiteStore) providerAccountRecord(ctx context.Context, accountID string) (*providerAccountRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at FROM proxy_runtime_provider_accounts WHERE account_id=?`, normalizeID(accountID))
	return scanSQLiteProviderAccount(row)
}

func scanSQLiteProviderAccount(row interface{ Scan(...any) error }) (*providerAccountRecord, error) {
	var record providerAccountRecord
	var enabled int
	var createdAt, updatedAt string
	err := row.Scan(&record.AccountID, &record.ProviderID, &record.DynamicProviderID, &record.DisplayName, &enabled, &record.CredentialSecret, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	record.Enabled = enabled != 0
	record.CreatedAt = parseSQLiteTime(createdAt)
	record.UpdatedAt = parseSQLiteTime(updatedAt)
	return &record, err
}

func existingProviderID(record *providerAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.ProviderID
}

func existingDynamicProviderID(record *providerAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.DynamicProviderID
}

func existingCredentialSecret(record *providerAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.CredentialSecret
}
