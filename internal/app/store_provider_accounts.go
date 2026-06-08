package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/random"
	"github.com/byte-v-forge/proxy-runtime/internal/secretbox"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PostgresStore) ListProviderAccounts(ctx context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+providerAccountColumns()+` FROM proxy_runtime_provider_accounts ORDER BY account_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxyruntimev1.ProxyProviderAccount{}
	for rows.Next() {
		record, err := scanProviderAccount(rows)
		if err != nil {
			return nil, err
		}
		account, err := s.providerAccountToProto(ctx, record)
		if err != nil {
			return nil, err
		}
		out = append(out, account)
	}
	return out, rows.Err()
}

func (s *PostgresStore) UpsertProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error) {
	accountID := normalizeID(req.GetAccountId())
	if accountID == "" {
		generated, err := generatedID("dynacct")
		if err != nil {
			return nil, err
		}
		accountID = generated
	}
	existing, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		existing = nil
	}
	providerID := strings.TrimSpace(req.GetProviderId())
	if providerID == "" && existing != nil {
		providerID = existing.ProviderID
	}
	providerID = firstNonEmpty(providerID, accountproxy.ProviderTen24)
	if !s.accountProviders.IsSupported(providerID) {
		return nil, fmt.Errorf("unsupported provider_id %q", providerID)
	}
	dynamicProviderID := runtimeSafeID(req.GetDynamicProviderId())
	if dynamicProviderID == "" && existing != nil {
		dynamicProviderID = existing.DynamicProviderID
	}
	secret := ""
	if existing != nil {
		secret = existing.CredentialSecret
	}
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
		ref, err := s.WriteSecret(ctx, secretref.WriteRequest{
			SecretID: secretref.StableID("proxy-runtime-provider-account-password", accountID),
			Provider: "proxy-runtime",
			Purpose:  "dynamic_ip_provider_password",
			Value:    rawPassword,
		})
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
	row := s.pool.QueryRow(ctx, `
INSERT INTO proxy_runtime_provider_accounts (account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (account_id) DO UPDATE SET provider_id=EXCLUDED.provider_id, dynamic_provider_id=EXCLUDED.dynamic_provider_id, display_name=EXCLUDED.display_name, enabled=EXCLUDED.enabled, credential_secret=EXCLUDED.credential_secret, updated_at=now()
RETURNING `+providerAccountColumns(), accountID, providerID, dynamicProviderID, displayName, enabled, secret)
	record, err := scanProviderAccount(row)
	if err != nil {
		return nil, err
	}
	return s.providerAccountToProto(ctx, record)
}

func credentialFromSecret(box secretbox.Box, secret string) *providerCredential {
	plain, err := box.Open(secret)
	if err != nil || len(plain) == 0 {
		return nil
	}
	var credential providerCredential
	if err := json.Unmarshal(plain, &credential); err != nil {
		return nil
	}
	return &credential
}

func generatedID(prefix string) (string, error) {
	suffix, err := random.Hex(6)
	if err != nil {
		return "", err
	}
	return prefix + "-" + suffix, nil
}

func (s *PostgresStore) DeleteProviderAccount(ctx context.Context, accountID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM proxy_runtime_provider_accounts WHERE account_id=$1`, normalizeID(accountID))
	return err
}

func (s *PostgresStore) ProviderAccount(ctx context.Context, accountID string) (*proxyruntimev1.ProxyProviderAccount, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return record.toProto(s.box), nil
}

func (s *PostgresStore) ProviderConfig(ctx context.Context, accountID string) (accountproxy.Config, string, error) {
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

func (s *PostgresStore) DefaultProviderAccountID(ctx context.Context) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT account_id FROM proxy_runtime_provider_accounts WHERE enabled ORDER BY updated_at DESC, account_id LIMIT 1`).Scan(&id)
	return id, err
}

func (s *PostgresStore) providerAccountRecord(ctx context.Context, accountID string) (*providerAccountRecord, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+providerAccountColumns()+` FROM proxy_runtime_provider_accounts WHERE account_id=$1`, normalizeID(accountID))
	return scanProviderAccount(row)
}

func providerAccountColumns() string {
	return `account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at`
}

func scanProviderAccount(row pgx.Row) (*providerAccountRecord, error) {
	var record providerAccountRecord
	err := row.Scan(&record.AccountID, &record.ProviderID, &record.DynamicProviderID, &record.DisplayName, &record.Enabled, &record.CredentialSecret, &record.CreatedAt, &record.UpdatedAt)
	return &record, err
}

func (s *PostgresStore) providerAccountToProto(ctx context.Context, record *providerAccountRecord) (*proxyruntimev1.ProxyProviderAccount, error) {
	return providerAccountToProto(ctx, s, s.box, record)
}

func (s *PostgresStore) ProviderAccountMutationState(ctx context.Context, accountID string) (providerAccountMutationState, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return providerAccountMutationState{}, err
	}
	credential := credentialFromSecret(s.box, record.CredentialSecret)
	state := providerAccountMutationState{
		ProviderID:         record.ProviderID,
		DynamicProviderID:  record.DynamicProviderID,
		PasswordConfigured: record.CredentialSecret != "",
	}
	if credential != nil {
		state.Username = credential.Username
		state.PasswordSecretRef = cloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password")
	}
	return state, nil
}

func providerConfigFromCredentialSecret(ctx context.Context, resolver secretref.Resolver, box secretbox.Box, providerID string, secret string) (accountproxy.Config, error) {
	credential := credentialFromSecret(box, secret)
	cfg := accountproxy.Config{ProviderID: providerID}
	if credential == nil {
		return cfg, nil
	}
	cfg.Username = credential.Username
	if password := credentialRawPassword(credential); password != "" {
		cfg.Password = password
		return cfg, nil
	}
	if ref := cloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password"); ref != nil {
		password, err := resolver.ResolveSecret(ctx, ref)
		if err != nil {
			return accountproxy.Config{}, err
		}
		cfg.Password = password
	}
	return cfg, nil
}

func providerAccountToProto(ctx context.Context, resolver secretref.Resolver, box secretbox.Box, record *providerAccountRecord) (*proxyruntimev1.ProxyProviderAccount, error) {
	account := record.toProto(box)
	credential := credentialFromSecret(box, record.CredentialSecret)
	if password := credentialRawPassword(credential); password != "" {
		account.PasswordValue = password
		return account, nil
	}
	if credential == nil || !secretRefConfigured(credential.PasswordSecretRef) {
		return account, nil
	}
	ref := cloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password")
	if ref == nil {
		return account, nil
	}
	password, err := resolver.ResolveSecret(ctx, ref)
	if err != nil {
		return account, nil
	}
	account.PasswordValue = password
	return account, nil
}

func credentialRawPassword(credential *providerCredential) string {
	if credential == nil {
		return ""
	}
	return firstNonEmpty(credential.PasswordValue, credential.Password)
}

func (r providerAccountRecord) toProto(box secretbox.Box) *proxyruntimev1.ProxyProviderAccount {
	status := proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_DISABLED
	if r.Enabled {
		status = proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED
	}
	credential := credentialFromSecret(box, r.CredentialSecret)
	username := ""
	if credential != nil {
		username = credential.Username
	}
	return &proxyruntimev1.ProxyProviderAccount{
		AccountId:            r.AccountID,
		ProviderId:           r.ProviderID,
		DynamicProviderId:    r.DynamicProviderID,
		DisplayName:          r.DisplayName,
		Status:               status,
		CredentialConfigured: r.CredentialSecret != "",
		CreatedAt:            timestamppb.New(r.CreatedAt),
		UpdatedAt:            timestamppb.New(r.UpdatedAt),
		Username:             username,
	}
}

func normalizeID(value string) string { return strings.TrimSpace(value) }
