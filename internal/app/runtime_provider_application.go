package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/jackc/pgx/v5"
)

type runtimeProviderApplication struct {
	runtime *Runtime
}

func newRuntimeProviderApplication(runtime *Runtime) runtimeProviderApplication {
	return runtimeProviderApplication{runtime: runtime}
}

func (s *RuntimeService) ListProxyProviders(ctx context.Context, _ *proxyruntimev1.ListProxyProvidersRequest) (*proxyruntimev1.ListProxyProvidersResponse, error) {
	return s.providers.ListProxyProviders(ctx)
}

func (s *RuntimeService) GetEgressGateway(ctx context.Context, _ *proxyruntimev1.GetEgressGatewayRequest) (*proxyruntimev1.GetEgressGatewayResponse, error) {
	return s.providers.GetEgressGateway(ctx)
}

func (s *RuntimeService) GetProxyPool(ctx context.Context, _ *proxyruntimev1.GetProxyPoolRequest) (*proxyruntimev1.GetProxyPoolResponse, error) {
	return s.providers.GetProxyPool(ctx)
}

func (s *RuntimeService) RefreshProxyPool(ctx context.Context, _ *proxyruntimev1.RefreshProxyPoolRequest) (*proxyruntimev1.RefreshProxyPoolResponse, error) {
	return s.providers.RefreshProxyPool(ctx)
}

func (s *RuntimeService) ListProxyProviderAccounts(ctx context.Context, _ *proxyruntimev1.ListProxyProviderAccountsRequest) (*proxyruntimev1.ListProxyProviderAccountsResponse, error) {
	return s.providers.ListProxyProviderAccounts(ctx)
}

func (s *RuntimeService) UpsertProxyProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.UpsertProxyProviderAccountResponse, error) {
	return s.providers.UpsertProxyProviderAccount(ctx, req)
}

func (s *RuntimeService) DeleteProxyProviderAccount(ctx context.Context, req *proxyruntimev1.DeleteProxyProviderAccountRequest) (*proxyruntimev1.DeleteProxyProviderAccountResponse, error) {
	return s.providers.DeleteProxyProviderAccount(ctx, req)
}

func (a runtimeProviderApplication) ListProxyProviders(ctx context.Context) (*proxyruntimev1.ListProxyProvidersResponse, error) {
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyProvidersResponse{Providers: a.runtime.accountProviders.Descriptors(dynamicIPGatewayMap(settings))}, nil
}

func (a runtimeProviderApplication) GetEgressGateway(ctx context.Context) (*proxyruntimev1.GetEgressGatewayResponse, error) {
	gateway, err := a.runtime.gateway(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetEgressGatewayResponse{Gateway: gateway}, nil
}

func (a runtimeProviderApplication) GetProxyPool(ctx context.Context) (*proxyruntimev1.GetProxyPoolResponse, error) {
	pool, err := a.runtime.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyPoolResponse{Pool: pool}, nil
}

func (a runtimeProviderApplication) RefreshProxyPool(ctx context.Context) (*proxyruntimev1.RefreshProxyPoolResponse, error) {
	if err := a.runtime.runReconcile(ctx); err != nil {
		return nil, unavailable("refresh proxy runtime failed", err)
	}
	pool, err := a.runtime.snapshot(ctx)
	if err != nil {
		return nil, internalError("", err)
	}
	return &proxyruntimev1.RefreshProxyPoolResponse{Pool: pool}, nil
}

func (a runtimeProviderApplication) ListProxyProviderAccounts(ctx context.Context) (*proxyruntimev1.ListProxyProviderAccountsResponse, error) {
	accounts, err := a.runtime.store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyProviderAccountsResponse{Accounts: accounts}, nil
}

func (a runtimeProviderApplication) UpsertProxyProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.UpsertProxyProviderAccountResponse, error) {
	if err := a.rejectActiveProviderAccountRuntimeMutation(ctx, req); err != nil {
		return nil, err
	}
	account, err := a.runtime.store.UpsertProviderAccount(ctx, req)
	if err != nil {
		return nil, invalidArgument("", err)
	}
	return &proxyruntimev1.UpsertProxyProviderAccountResponse{Account: account}, nil
}

func (a runtimeProviderApplication) DeleteProxyProviderAccount(ctx context.Context, req *proxyruntimev1.DeleteProxyProviderAccountRequest) (*proxyruntimev1.DeleteProxyProviderAccountResponse, error) {
	if err := a.deleteProviderAccount(ctx, req.GetAccountId()); err != nil {
		return nil, invalidArgument("", err)
	}
	return &proxyruntimev1.DeleteProxyProviderAccountResponse{}, nil
}

func (a runtimeProviderApplication) rejectActiveProviderAccountRuntimeMutation(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) error {
	providerAccountID := strings.TrimSpace(req.GetAccountId())
	if providerAccountID == "" {
		return nil
	}
	active, err := a.runtime.store.ProviderAccountHasBlockingLease(ctx, providerAccountID)
	if err != nil {
		return err
	}
	if !active {
		return nil
	}
	record, err := a.runtime.store.providerAccountRecord(ctx, providerAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if providerID := strings.TrimSpace(req.GetProviderId()); providerID != "" && providerID != record.ProviderID {
		return failedPrecondition("provider account has active leases", nil)
	}
	credential := credentialFromSecret(a.runtime.store.box, record.CredentialSecret)
	if username := strings.TrimSpace(req.GetUsername()); username != "" && (credential == nil || username != credential.Username) {
		return failedPrecondition("provider account has active leases", nil)
	}
	if req.GetClearPassword() || secretRefConfigured(req.GetPasswordSecretRef()) || strings.TrimSpace(req.GetPasswordValue()) != "" {
		return failedPrecondition("provider account has active leases", nil)
	}
	return nil
}

func (a runtimeProviderApplication) deleteProviderAccount(ctx context.Context, providerAccountID string) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return errors.New("provider account_id is required")
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		lock, err := a.runtime.leaseLocks.LockProviderAccount(ctx, providerAccountID)
		if err != nil {
			return err
		}
		leases, err := a.runtime.store.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID)
		if err != nil {
			_ = lock.Unlock(ctx)
			return fmt.Errorf("list blocking proxy leases for provider account %q: %w", providerAccountID, err)
		}
		if len(leases) == 0 {
			err = a.runtime.store.DeleteProviderAccount(ctx, providerAccountID)
			_ = lock.Unlock(ctx)
			return err
		}
		_ = lock.Unlock(ctx)
		for _, lease := range leases {
			if leaseCleanupPending(lease) {
				if err := a.runtime.leaseCoordinator.cleanupPendingLeaseFact(ctx, lease); err != nil {
					return fmt.Errorf("cleanup proxy lease %q for provider account %q: %w", lease.GetLeaseId(), providerAccountID, err)
				}
				continue
			}
			if _, err := a.runtime.leaseCoordinator.releaseLease(ctx, &proxyruntimev1.ReleaseProxyLeaseRequest{LeaseId: lease.GetLeaseId(), AccountId: lease.GetAccountId(), Purpose: lease.GetPurpose()}); err != nil {
				return fmt.Errorf("release blocking proxy lease %q for provider account %q: %w", lease.GetLeaseId(), providerAccountID, err)
			}
		}
	}
}
