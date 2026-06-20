package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

func (s *Store) ListProviderAccounts(ctx context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error) {
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
		account, err := store.ProviderAccountToProto(ctx, s, s.box, record)
		if err != nil {
			return nil, err
		}
		out = append(out, account)
	}
	return out, rows.Err()
}

func (s *Store) UpsertProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error) {
	accountID := store.NormalizeID(req.GetAccountId())
	if accountID == "" {
		generated, err := store.GeneratedID("dynacct")
		if err != nil {
			return nil, err
		}
		accountID = generated
	}
	existing, err := s.providerAccountRecord(ctx, accountID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	providerID := appcore.FirstNonEmpty(strings.TrimSpace(req.GetProviderId()), existingProviderID(existing), s.accountProviders.DefaultProviderID())
	if providerID == "" {
		return nil, appcore.FailedPrecondition("provider_id is required", nil)
	}
	if !s.accountProviders.IsSupported(providerID) {
		return nil, fmt.Errorf("unsupported provider_id %q", providerID)
	}
	dynamicProviderID := appcore.FirstNonEmpty(appcore.RuntimeSafeID(req.GetDynamicProviderId()), existingDynamicProviderID(existing))
	secret := existingCredentialSecret(existing)
	credential := store.ProviderCredential{}
	if current := store.CredentialFromSecret(s.box, secret); current != nil {
		credential = *current
	}
	if req.GetClearPassword() {
		credential.PasswordSecretRef = nil
	}
	if username := strings.TrimSpace(req.GetUsername()); username != "" {
		credential.Username = username
	}
	if ref := appcore.CloneSecretRef(req.GetPasswordSecretRef(), "proxy-runtime", "dynamic_ip_provider_password"); ref != nil {
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
	displayName := appcore.FirstNonEmpty(req.GetDisplayName(), accountID)
	if enabled {
		cfg, err := store.ProviderConfigFromCredentialSecret(ctx, s, s.box, providerID, secret)
		if err != nil {
			return nil, fmt.Errorf("enabled provider account invalid: %w", err)
		}
		if err := s.accountProviders.Validate(cfg); err != nil {
			return nil, fmt.Errorf("enabled provider account invalid: %w", err)
		}
	}
	now := s.clock.Now().UTC()
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
	return store.ProviderAccountToProto(ctx, s, s.box, record)
}

func (s *Store) DeleteProviderAccount(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM proxy_runtime_provider_accounts WHERE account_id=?`, store.NormalizeID(accountID))
	return err
}

func (s *Store) ProviderAccount(ctx context.Context, accountID string) (*proxyruntimev1.ProxyProviderAccount, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return store.ProviderAccountToProto(ctx, s, s.box, record)
}

func (s *Store) ProviderAccountMutationState(ctx context.Context, accountID string) (store.ProviderAccountMutationState, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return store.ProviderAccountMutationState{}, err
	}
	credential := store.CredentialFromSecret(s.box, record.CredentialSecret)
	state := store.ProviderAccountMutationState{ProviderID: record.ProviderID, DynamicProviderID: record.DynamicProviderID, PasswordConfigured: record.CredentialSecret != ""}
	if credential != nil {
		state.Username = credential.Username
		state.PasswordSecretRef = appcore.CloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password")
	}
	return state, nil
}

func (s *Store) ProviderConfig(ctx context.Context, accountID string) (accountproxy.Config, string, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return accountproxy.Config{}, "", err
	}
	if !record.Enabled {
		return accountproxy.Config{}, "", errors.New("provider account is disabled")
	}
	cfg, err := store.ProviderConfigFromCredentialSecret(ctx, s, s.box, record.ProviderID, record.CredentialSecret)
	if err != nil {
		return accountproxy.Config{}, "", err
	}
	return cfg, record.AccountID, s.accountProviders.Validate(cfg)
}

func (s *Store) DefaultProviderAccountID(ctx context.Context) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT account_id FROM proxy_runtime_provider_accounts WHERE enabled=1 ORDER BY updated_at DESC, account_id LIMIT 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	return id, err
}

func (s *Store) providerAccountRecord(ctx context.Context, accountID string) (*store.ProviderAccountRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at FROM proxy_runtime_provider_accounts WHERE account_id=?`, store.NormalizeID(accountID))
	return scanSQLiteProviderAccount(row)
}

func scanSQLiteProviderAccount(row interface{ Scan(...any) error }) (*store.ProviderAccountRecord, error) {
	var record store.ProviderAccountRecord
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

func existingProviderID(record *store.ProviderAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.ProviderID
}

func existingDynamicProviderID(record *store.ProviderAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.DynamicProviderID
}

func existingCredentialSecret(record *store.ProviderAccountRecord) string {
	if record == nil {
		return ""
	}
	return record.CredentialSecret
}
