package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

const providerAccountDeleteTimeout = 2 * time.Minute

type runtimeProviderRepository interface {
	ListProviderAccounts(context.Context) ([]*proxyruntimev1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxyruntimev1.ProxyProviderAccount, error)
	ProviderAccountHasBlockingLease(context.Context, string) (bool, error)
	BlockingLeaseFactsByProviderAccount(context.Context, string) ([]*proxyruntimev1.ProxyDynamicLease, error)
	ProviderAccountMutationState(context.Context, string) (providerAccountMutationState, error)
}

type runtimeProviderSettings interface {
	load(context.Context) (*runtimeSettingsFile, error)
}

type runtimeProviderLeaseOperations interface {
	CleanupPendingLeaseFact(context.Context, *proxyruntimev1.ProxyDynamicLease) error
	ReleaseProxyLease(context.Context, *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ReleaseProxyLeaseResponse, error)
}

type runtimeProviderDescriptorsFunc func(map[string][]accountproxy.Gateway) []*proxyruntimev1.ProxyProviderDescriptor

type runtimeProviderApplicationDependencies struct {
	Store               runtimeProviderRepository
	Settings            runtimeProviderSettings
	ProviderDescriptors runtimeProviderDescriptorsFunc
	Locks               leaseapp.LockManager
	LeaseOperations     func() runtimeProviderLeaseOperations
	Logger              leaseapp.Logger
}

type runtimeProviderApplication struct {
	store               runtimeProviderRepository
	settings            runtimeProviderSettings
	providerDescriptors runtimeProviderDescriptorsFunc
	locks               leaseapp.LockManager
	leases              func() runtimeProviderLeaseOperations
	logger              leaseapp.Logger
}

func newRuntimeProviderApplication(deps runtimeProviderApplicationDependencies) runtimeProviderApplication {
	return runtimeProviderApplication{
		store:               deps.Store,
		settings:            deps.Settings,
		providerDescriptors: deps.ProviderDescriptors,
		locks:               deps.Locks,
		leases:              deps.LeaseOperations,
		logger:              deps.Logger,
	}
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
	settingsStore, err := a.requireSettings()
	if err != nil {
		return nil, err
	}
	descriptors, err := a.requireProviderDescriptors()
	if err != nil {
		return nil, err
	}
	settings, err := settingsStore.load(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyProvidersResponse{Providers: descriptors(dynamicIPEndpointMap(settings))}, nil
}

func (a runtimeProviderApplication) ListProxyProviderAccounts(ctx context.Context) (*proxyruntimev1.ListProxyProviderAccountsResponse, error) {
	store, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	accounts, err := store.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.ListProxyProviderAccountsResponse{Accounts: accounts}, nil
}

func (a runtimeProviderApplication) UpsertProxyProviderAccount(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) (*proxyruntimev1.UpsertProxyProviderAccountResponse, error) {
	store, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	if err := a.rejectActiveProviderAccountRuntimeMutation(ctx, req); err != nil {
		return nil, err
	}
	if err := a.normalizeProviderAccountDynamicProvider(ctx, req); err != nil {
		return nil, invalidArgument("", err)
	}
	account, err := store.UpsertProviderAccount(ctx, req)
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
	settingsStore, err := a.requireSettings()
	if err != nil {
		return err
	}
	settings, err := settingsStore.load(ctx)
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
	providerAccountID := strings.TrimSpace(req.GetAccountId())
	if providerAccountID == "" {
		return nil, invalidArgument("provider account_id is required", nil)
	}
	store, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	if _, err := store.ProviderAccount(ctx, providerAccountID); err != nil {
		return nil, invalidArgument("provider account is not configured", err)
	}
	a.deleteProviderAccountInBackground(providerAccountID)
	return &proxyruntimev1.DeleteProxyProviderAccountResponse{}, nil
}

func (a runtimeProviderApplication) rejectActiveProviderAccountRuntimeMutation(ctx context.Context, req *proxyruntimev1.UpsertProxyProviderAccountRequest) error {
	store, err := a.requireStore()
	if err != nil {
		return err
	}
	providerAccountID := strings.TrimSpace(req.GetAccountId())
	if providerAccountID == "" {
		return nil
	}
	active, err := store.ProviderAccountHasBlockingLease(ctx, providerAccountID)
	if err != nil {
		return err
	}
	if !active {
		return nil
	}
	state, err := store.ProviderAccountMutationState(ctx, providerAccountID)
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
	store, err := a.requireStore()
	if err != nil {
		return err
	}
	leaseOperations, err := a.requireLeaseOperations()
	if err != nil {
		return err
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var leases []*proxyruntimev1.ProxyDynamicLease
		deleted := false
		err := a.withProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
			var err error
			leases, err = store.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID)
			if err != nil {
				return fmt.Errorf("list blocking proxy leases for provider account %q: %w", providerAccountID, err)
			}
			if len(leases) == 0 {
				deleted = true
				return store.DeleteProviderAccount(ctx, providerAccountID)
			}
			return nil
		})
		if err != nil || deleted {
			return err
		}
		for _, lease := range leases {
			if leaseapp.CleanupPending(lease) {
				if err := leaseOperations.CleanupPendingLeaseFact(ctx, lease); err != nil {
					return fmt.Errorf("cleanup proxy lease %q for provider account %q: %w", lease.GetLeaseId(), providerAccountID, err)
				}
				continue
			}
			if _, err := leaseOperations.ReleaseProxyLease(ctx, &proxyruntimev1.ReleaseProxyLeaseRequest{LeaseId: lease.GetLeaseId(), AccountId: lease.GetAccountId(), Purpose: lease.GetPurpose()}); err != nil {
				return fmt.Errorf("release blocking proxy lease %q for provider account %q: %w", lease.GetLeaseId(), providerAccountID, err)
			}
		}
	}
}

func (a runtimeProviderApplication) requireStore() (runtimeProviderRepository, error) {
	if a.store == nil {
		return nil, internalError("provider account repository is not configured", nil)
	}
	return a.store, nil
}

func (a runtimeProviderApplication) requireSettings() (runtimeProviderSettings, error) {
	if a.settings == nil {
		return nil, internalError("provider settings repository is not configured", nil)
	}
	return a.settings, nil
}

func (a runtimeProviderApplication) requireProviderDescriptors() (runtimeProviderDescriptorsFunc, error) {
	if a.providerDescriptors == nil {
		return nil, internalError("provider descriptor registry is not configured", nil)
	}
	return a.providerDescriptors, nil
}

func (a runtimeProviderApplication) requireLeaseOperations() (runtimeProviderLeaseOperations, error) {
	if a.leases == nil {
		return nil, internalError("lease application is not configured", nil)
	}
	operations := a.leases()
	if operations == nil {
		return nil, internalError("lease application is not configured", nil)
	}
	return operations, nil
}

func (a runtimeProviderApplication) withProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	if a.locks == nil {
		return fn(ctx)
	}
	return a.locks.WithProviderAccountLock(ctx, providerAccountID, fn)
}

func (a runtimeProviderApplication) warn(message string, args ...any) {
	if a.logger != nil {
		a.logger.Warn(message, args...)
	}
}

func (a runtimeProviderApplication) info(message string, args ...any) {
	if a.logger != nil {
		a.logger.Info(message, args...)
	}
}

func (a runtimeProviderApplication) deleteProviderAccountInBackground(providerAccountID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), providerAccountDeleteTimeout)
		defer cancel()
		if err := a.deleteProviderAccount(ctx, providerAccountID); err != nil {
			a.warn("delete provider account failed", "provider_account_id", providerAccountID, "error", err)
			return
		}
		a.info("delete provider account finished", "provider_account_id", providerAccountID)
	}()
}
