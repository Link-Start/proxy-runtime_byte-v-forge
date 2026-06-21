package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/byte-v-forge/proxy-gateway/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/dynamic"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

const providerAccountDeleteTimeout = 2 * time.Minute

type Repository interface {
	ListProviderAccounts(context.Context) ([]*proxygatewayv1.ProxyProviderAccount, error)
	UpsertProviderAccount(context.Context, *proxygatewayv1.UpsertProxyProviderAccountRequest) (*proxygatewayv1.ProxyProviderAccount, error)
	DeleteProviderAccount(context.Context, string) error
	ProviderAccount(context.Context, string) (*proxygatewayv1.ProxyProviderAccount, error)
	ProviderAccountHasBlockingLease(context.Context, string) (bool, error)
	BlockingLeaseFactsByProviderAccount(context.Context, string, int) ([]*proxygatewayv1.ProxyDynamicLease, error)
	ProviderAccountMutationState(context.Context, string) (store.ProviderAccountMutationState, error)
}

type LeaseOperations interface {
	CleanupPendingLeaseFact(context.Context, *proxygatewayv1.ProxyDynamicLease) error
	ReleaseProxyLease(context.Context, *proxygatewayv1.ReleaseProxyLeaseRequest) (*proxygatewayv1.ReleaseProxyLeaseResponse, error)
}

type DescriptorsFunc func(map[string][]accountproxy.Gateway) []*proxygatewayv1.ProxyProviderDescriptor

type Dependencies struct {
	Store               Repository
	LoadSettings        func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	ProviderDescriptors DescriptorsFunc
	Locks               leaseapp.LockManager
	LeaseOperations     func() LeaseOperations
	Logger              leaseapp.Logger
}

type Service struct {
	store               Repository
	loadSettings        func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	providerDescriptors DescriptorsFunc
	locks               leaseapp.LockManager
	leases              func() LeaseOperations
	logger              leaseapp.Logger
}

func New(deps Dependencies) Service {
	return Service{
		store:               deps.Store,
		loadSettings:        deps.LoadSettings,
		providerDescriptors: deps.ProviderDescriptors,
		locks:               deps.Locks,
		leases:              deps.LeaseOperations,
		logger:              deps.Logger,
	}
}

func (a Service) ListProxyProviders(ctx context.Context) (*proxygatewayv1.ListProxyProvidersResponse, error) {
	loadSettings, err := a.requireSettings()
	if err != nil {
		return nil, err
	}
	descriptors, err := a.requireProviderDescriptors()
	if err != nil {
		return nil, err
	}
	settings, err := loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.ListProxyProvidersResponse{Providers: descriptors(dynamic.EndpointMap(settings))}, nil
}

func (a Service) ListProxyProviderAccounts(ctx context.Context) (*proxygatewayv1.ListProxyProviderAccountsResponse, error) {
	repo, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	accounts, err := repo.ListProviderAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.ListProxyProviderAccountsResponse{Accounts: accounts}, nil
}

func (a Service) UpsertProxyProviderAccount(ctx context.Context, req *proxygatewayv1.UpsertProxyProviderAccountRequest) (*proxygatewayv1.UpsertProxyProviderAccountResponse, error) {
	repo, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	if err := a.rejectActiveProviderAccountRuntimeMutation(ctx, req); err != nil {
		return nil, err
	}
	if err := a.normalizeProviderAccountDynamicProvider(ctx, req); err != nil {
		return nil, appcore.InvalidArgument("", err)
	}
	account, err := repo.UpsertProviderAccount(ctx, req)
	if err != nil {
		return nil, appcore.InvalidArgument("", err)
	}
	return &proxygatewayv1.UpsertProxyProviderAccountResponse{Account: account}, nil
}

func (a Service) normalizeProviderAccountDynamicProvider(ctx context.Context, req *proxygatewayv1.UpsertProxyProviderAccountRequest) error {
	dynamicProviderID := appcore.RuntimeSafeID(req.GetDynamicProviderId())
	if dynamicProviderID == "" {
		return nil
	}
	loadSettings, err := a.requireSettings()
	if err != nil {
		return err
	}
	settings, err := loadSettings(ctx)
	if err != nil {
		return err
	}
	for _, provider := range dynamic.ProviderInstances(settings) {
		if provider.DynamicProviderID != dynamicProviderID {
			continue
		}
		if providerID := strings.TrimSpace(req.GetProviderId()); providerID != "" && providerID != provider.ProviderID {
			return fmt.Errorf("dynamic provider %q uses provider_id %q", dynamicProviderID, provider.ProviderID)
		}
		req.DynamicProviderId = dynamicProviderID
		req.ProviderId = provider.ProviderID
		return nil
	}
	return fmt.Errorf("dynamic provider %q is not enabled", dynamicProviderID)
}

func (a Service) DeleteProxyProviderAccount(ctx context.Context, req *proxygatewayv1.DeleteProxyProviderAccountRequest) (*proxygatewayv1.DeleteProxyProviderAccountResponse, error) {
	providerAccountID := strings.TrimSpace(req.GetAccountId())
	if providerAccountID == "" {
		return nil, appcore.InvalidArgument("provider account_id is required", nil)
	}
	repo, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	if _, err := repo.ProviderAccount(ctx, providerAccountID); err != nil {
		return nil, appcore.InvalidArgument("provider account is not configured", err)
	}
	a.deleteProviderAccountInBackground(providerAccountID)
	return &proxygatewayv1.DeleteProxyProviderAccountResponse{}, nil
}

func (a Service) rejectActiveProviderAccountRuntimeMutation(ctx context.Context, req *proxygatewayv1.UpsertProxyProviderAccountRequest) error {
	repo, err := a.requireStore()
	if err != nil {
		return err
	}
	providerAccountID := strings.TrimSpace(req.GetAccountId())
	if providerAccountID == "" {
		return nil
	}
	active, err := repo.ProviderAccountHasBlockingLease(ctx, providerAccountID)
	if err != nil {
		return err
	}
	if !active {
		return nil
	}
	state, err := repo.ProviderAccountMutationState(ctx, providerAccountID)
	if err != nil {
		if store.IsNotFound(err) {
			return nil
		}
		return err
	}
	if providerID := strings.TrimSpace(req.GetProviderId()); providerID != "" && providerID != state.ProviderID {
		return appcore.FailedPrecondition("provider account has active leases", nil)
	}
	if dynamicProviderID := appcore.RuntimeSafeID(req.GetDynamicProviderId()); dynamicProviderID != "" && dynamicProviderID != state.DynamicProviderID {
		return appcore.FailedPrecondition("provider account has active leases", nil)
	}
	if username := strings.TrimSpace(req.GetUsername()); username != "" && username != state.Username {
		return appcore.FailedPrecondition("provider account has active leases", nil)
	}
	if req.GetClearPassword() || appcore.SecretRefConfigured(req.GetPasswordSecretRef()) || strings.TrimSpace(req.GetPasswordValue()) != "" {
		return appcore.FailedPrecondition("provider account has active leases", nil)
	}
	return nil
}

func (a Service) deleteProviderAccount(ctx context.Context, providerAccountID string) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return appcore.InvalidArgument("provider account_id is required", nil)
	}
	repo, err := a.requireStore()
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
		var leases []*proxygatewayv1.ProxyDynamicLease
		deleted := false
		err := a.withProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
			var err error
			leases, err = repo.BlockingLeaseFactsByProviderAccount(ctx, providerAccountID, kernel.DefaultBlockingLeaseFactLimit)
			if err != nil {
				return fmt.Errorf("list blocking proxy leases for provider account %q: %w", providerAccountID, err)
			}
			if len(leases) == 0 {
				deleted = true
				return repo.DeleteProviderAccount(ctx, providerAccountID)
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
			if _, err := leaseOperations.ReleaseProxyLease(ctx, &proxygatewayv1.ReleaseProxyLeaseRequest{LeaseId: lease.GetLeaseId(), AccountId: lease.GetAccountId(), Purpose: lease.GetPurpose()}); err != nil {
				return fmt.Errorf("release blocking proxy lease %q for provider account %q: %w", lease.GetLeaseId(), providerAccountID, err)
			}
		}
	}
}

func (a Service) requireStore() (Repository, error) {
	if a.store == nil {
		return nil, appcore.InternalError("provider account repository is not configured", nil)
	}
	return a.store, nil
}

func (a Service) requireSettings() (func(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error), error) {
	if a.loadSettings == nil {
		return nil, appcore.InternalError("provider settings repository is not configured", nil)
	}
	return a.loadSettings, nil
}

func (a Service) requireProviderDescriptors() (DescriptorsFunc, error) {
	if a.providerDescriptors == nil {
		return nil, appcore.InternalError("provider descriptor registry is not configured", nil)
	}
	return a.providerDescriptors, nil
}

func (a Service) requireLeaseOperations() (LeaseOperations, error) {
	if a.leases == nil {
		return nil, appcore.InternalError("lease application is not configured", nil)
	}
	operations := a.leases()
	if operations == nil {
		return nil, appcore.InternalError("lease application is not configured", nil)
	}
	return operations, nil
}

func (a Service) withProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	if a.locks == nil {
		return fn(ctx)
	}
	return a.locks.WithProviderAccountLock(ctx, providerAccountID, fn)
}

func (a Service) warn(message string, args ...any) {
	if a.logger != nil {
		a.logger.Warn(message, args...)
	}
}

func (a Service) info(message string, args ...any) {
	if a.logger != nil {
		a.logger.Info(message, args...)
	}
}

func (a Service) deleteProviderAccountInBackground(providerAccountID string) {
	go func() {
		startedAt := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), providerAccountDeleteTimeout)
		defer cancel()
		if err := a.deleteProviderAccount(ctx, providerAccountID); err != nil {
			a.warn("delete provider account failed", "provider_account_id", providerAccountID, "duration_ms", time.Since(startedAt).Milliseconds(), "error_type", appcore.ErrorLogType(err))
			return
		}
		a.info("delete provider account finished", "provider_account_id", providerAccountID, "duration_ms", time.Since(startedAt).Milliseconds())
	}()
}
