package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-gateway/internal/secretref"
	"github.com/jackc/pgx/v5"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) ListProviderAccounts(ctx context.Context) ([]*proxygatewayv1.ProxyProviderAccount, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+providerAccountColumns()+` FROM proxy_gateway_provider_accounts ORDER BY account_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*proxygatewayv1.ProxyProviderAccount{}
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

func (s *Store) UpsertProviderAccount(ctx context.Context, req *proxygatewayv1.UpsertProxyProviderAccountRequest) (*proxygatewayv1.ProxyProviderAccount, error) {
	accountID := store.NormalizeID(req.GetAccountId())
	if accountID == "" {
		generated, err := store.GeneratedID("dynacct")
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
	providerID = appcore.FirstNonEmpty(providerID, s.accountProviders.DefaultProviderID())
	if providerID == "" {
		return nil, appcore.FailedPrecondition("provider_id is required", nil)
	}
	if !s.accountProviders.IsSupported(providerID) {
		return nil, fmt.Errorf("unsupported provider_id %q", providerID)
	}
	dynamicProviderID := appcore.RuntimeSafeID(req.GetDynamicProviderId())
	if dynamicProviderID == "" && existing != nil {
		dynamicProviderID = existing.DynamicProviderID
	}
	secret := ""
	if existing != nil {
		secret = existing.CredentialSecret
	}
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
	if ref := appcore.CloneSecretRef(req.GetPasswordSecretRef(), "proxy-gateway", "dynamic_ip_provider_password"); ref != nil {
		credential.PasswordSecretRef = ref
	}
	if rawPassword := strings.TrimSpace(req.GetPasswordValue()); rawPassword != "" {
		ref, err := s.WriteSecret(ctx, secretref.WriteRequest{
			SecretID: secretref.StableID("proxy-gateway-provider-account-password", accountID),
			Provider: "proxy-gateway",
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
	row := s.pool.QueryRow(ctx, `
INSERT INTO proxy_gateway_provider_accounts (account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (account_id) DO UPDATE SET provider_id=EXCLUDED.provider_id, dynamic_provider_id=EXCLUDED.dynamic_provider_id, display_name=EXCLUDED.display_name, enabled=EXCLUDED.enabled, credential_secret=EXCLUDED.credential_secret, updated_at=now()
RETURNING `+providerAccountColumns(), accountID, providerID, dynamicProviderID, displayName, enabled, secret)
	record, err := scanProviderAccount(row)
	if err != nil {
		return nil, err
	}
	return s.providerAccountToProto(ctx, record)
}

func (s *Store) DeleteProviderAccount(ctx context.Context, accountID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM proxy_gateway_provider_accounts WHERE account_id=$1`, store.NormalizeID(accountID))
	return err
}

func (s *Store) ProviderAccount(ctx context.Context, accountID string) (*proxygatewayv1.ProxyProviderAccount, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return record.ToProto(s.box), nil
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
	err := s.pool.QueryRow(ctx, `SELECT account_id FROM proxy_gateway_provider_accounts WHERE enabled ORDER BY updated_at DESC, account_id LIMIT 1`).Scan(&id)
	return id, err
}

func (s *Store) providerAccountRecord(ctx context.Context, accountID string) (*store.ProviderAccountRecord, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+providerAccountColumns()+` FROM proxy_gateway_provider_accounts WHERE account_id=$1`, store.NormalizeID(accountID))
	return scanProviderAccount(row)
}

func providerAccountColumns() string {
	return `account_id, provider_id, dynamic_provider_id, display_name, enabled, credential_secret, created_at, updated_at`
}

func scanProviderAccount(row pgx.Row) (*store.ProviderAccountRecord, error) {
	var record store.ProviderAccountRecord
	err := row.Scan(&record.AccountID, &record.ProviderID, &record.DynamicProviderID, &record.DisplayName, &record.Enabled, &record.CredentialSecret, &record.CreatedAt, &record.UpdatedAt)
	return &record, err
}

func (s *Store) providerAccountToProto(ctx context.Context, record *store.ProviderAccountRecord) (*proxygatewayv1.ProxyProviderAccount, error) {
	return store.ProviderAccountToProto(ctx, s, s.box, record)
}

func (s *Store) ProviderAccountMutationState(ctx context.Context, accountID string) (store.ProviderAccountMutationState, error) {
	record, err := s.providerAccountRecord(ctx, accountID)
	if err != nil {
		return store.ProviderAccountMutationState{}, err
	}
	credential := store.CredentialFromSecret(s.box, record.CredentialSecret)
	state := store.ProviderAccountMutationState{
		ProviderID:         record.ProviderID,
		DynamicProviderID:  record.DynamicProviderID,
		PasswordConfigured: record.CredentialSecret != "",
	}
	if credential != nil {
		state.Username = credential.Username
		state.PasswordSecretRef = appcore.CloneSecretRef(credential.PasswordSecretRef, "proxy-gateway", "dynamic_ip_provider_password")
	}
	return state, nil
}
