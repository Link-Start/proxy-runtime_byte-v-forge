package store

import (
	"context"
	"encoding/json"
	"time"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/random"
	"github.com/byte-v-forge/proxy-runtime/internal/secretbox"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

type ProviderCredential struct {
	Username          string              `json:"username"`
	Password          string              `json:"password,omitempty"`
	PasswordValue     string              `json:"password_value,omitempty"`
	PasswordSecretRef *commonv1.SecretRef `json:"password_secret_ref,omitempty"`
}

type ProviderAccountRecord struct {
	AccountID         string
	ProviderID        string
	DynamicProviderID string
	DisplayName       string
	Enabled           bool
	CredentialSecret  string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func CredentialFromSecret(box secretbox.Box, secret string) *ProviderCredential {
	plain, err := box.Open(secret)
	if err != nil || len(plain) == 0 {
		return nil
	}
	var credential ProviderCredential
	if err := json.Unmarshal(plain, &credential); err != nil {
		return nil
	}
	return &credential
}

func GeneratedID(prefix string) (string, error) {
	suffix, err := random.Hex(6)
	if err != nil {
		return "", err
	}
	return prefix + "-" + suffix, nil
}

func ProviderConfigFromCredentialSecret(ctx context.Context, resolver secretref.Resolver, box secretbox.Box, providerID string, secret string) (accountproxy.Config, error) {
	credential := CredentialFromSecret(box, secret)
	cfg := accountproxy.Config{ProviderID: providerID}
	if credential == nil {
		return cfg, nil
	}
	cfg.Username = credential.Username
	if password := credentialRawPassword(credential); password != "" {
		cfg.Password = password
		return cfg, nil
	}
	if ref := appcore.CloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password"); ref != nil {
		password, err := resolver.ResolveSecret(ctx, ref)
		if err != nil {
			return accountproxy.Config{}, err
		}
		cfg.Password = password
	}
	return cfg, nil
}

func ProviderAccountToProto(ctx context.Context, resolver secretref.Resolver, box secretbox.Box, record *ProviderAccountRecord) (*proxyruntimev1.ProxyProviderAccount, error) {
	account := record.ToProto(box)
	credential := CredentialFromSecret(box, record.CredentialSecret)
	if password := credentialRawPassword(credential); password != "" {
		account.PasswordValue = password
		return account, nil
	}
	if credential == nil || !appcore.SecretRefConfigured(credential.PasswordSecretRef) {
		return account, nil
	}
	ref := appcore.CloneSecretRef(credential.PasswordSecretRef, "proxy-runtime", "dynamic_ip_provider_password")
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

func credentialRawPassword(credential *ProviderCredential) string {
	if credential == nil {
		return ""
	}
	return appcore.FirstNonEmpty(credential.PasswordValue, credential.Password)
}

func (r ProviderAccountRecord) ToProto(box secretbox.Box) *proxyruntimev1.ProxyProviderAccount {
	status := proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_DISABLED
	if r.Enabled {
		status = proxyruntimev1.ProxyProviderAccountStatus_PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED
	}
	credential := CredentialFromSecret(box, r.CredentialSecret)
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
