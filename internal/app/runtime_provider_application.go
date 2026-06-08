package app

import (
	"context"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
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
	return &proxyruntimev1.ListProxyProvidersResponse{Providers: a.runtime.accountProviders.Descriptors(dynamicIPEndpointMap(settings))}, nil
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
	if err := a.normalizeProviderAccountDynamicProvider(ctx, req); err != nil {
		return nil, invalidArgument("", err)
	}
	account, err := a.runtime.store.UpsertProviderAccount(ctx, req)
	if err != nil {
		return nil, invalidArgument("", err)
	}
	return &proxyruntimev1.UpsertProxyProviderAccountResponse{Account: account}, nil
}

func (a runtimeProviderApplication) normalizeProviderAccountDynamicProvider(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) error {
	dynamicProviderID := runtimeSafeID(req.GetDynamicProviderId())
	if dynamicProviderID == "" {
		return nil
	}
	settings, err := a.runtime.settings.load(ctx)
	if err != nil {
		return err
	}
	for _, provider := range dynamicIPProviderInstances(settings) {
		if provider.dynamicProviderID != dynamicProviderID {
			continue
		}
		if providerID := strings.TrimSpace(req.GetProviderId()); providerID != "" && providerID != provider.providerID {
			return fmt.Errorf("dynamic provider %q uses provider_id %q", dynamicProviderID, provider.providerID)
		}
		req.DynamicProviderId = dynamicProviderID
		req.ProviderId = provider.providerID
		return nil
	}
	return fmt.Errorf("dynamic provider %q is not enabled", dynamicProviderID)
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
	state, err := a.runtime.store.ProviderAccountMutationState(ctx, providerAccountID)
	if err != nil {
		if isStoreNotFound(err) {
			return nil
		}
		return err
	}
	if providerID := strings.TrimSpace(req.GetProviderId()); providerID != "" && providerID != state.ProviderID {
		return failedPrecondition("provider account has active leases", nil)
	}
	if dynamicProviderID := runtimeSafeID(req.GetDynamicProviderId()); dynamicProviderID != "" && dynamicProviderID != state.DynamicProviderID {
		return failedPrecondition("provider account has active leases", nil)
	}
	if username := strings.TrimSpace(req.GetUsername()); username != "" && username != state.Username {
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
		return invalidArgument("provider account_id is required", nil)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var leases []*proxyruntimev1.ProxyDynamicLease
		deleted := false
		err := a.runtime.leaseLocks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
			var err error
			leases, err = a.runtime.store.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID)
			if err != nil {
				return fmt.Errorf("list blocking proxy leases for provider account %q: %w", providerAccountID, err)
			}
			if len(leases) == 0 {
				deleted = true
				return a.runtime.store.DeleteProviderAccount(ctx, providerAccountID)
			}
			return nil
		})
		if err != nil || deleted {
			return err
		}
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
