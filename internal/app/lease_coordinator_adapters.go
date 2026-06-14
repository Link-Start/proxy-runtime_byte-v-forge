package app

import (
	"context"
	"fmt"
	"net/http"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/random"
)

const leaseIDByteLength = 12

type leaseRegistrySessionProviderFactory struct {
	registry *providerregistry.Registry
	client   *http.Client
}

func (f leaseRegistrySessionProviderFactory) NewSessionProvider(providerCfg accountproxy.Config) (provider.SessionProvider, error) {
	if f.registry == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return f.registry.NewSessionProvider(providerCfg, f.client)
}

type leaseRuntimeLockManager struct {
	locks leaseRuntimeLocks
}

func (m leaseRuntimeLockManager) WithAccountLock(ctx context.Context, accountID string, fn leaseapp.LockFunc) error {
	return m.locks.WithAccountLock(ctx, accountID, func(ctx context.Context) error {
		return fn(ctx)
	})
}

func (m leaseRuntimeLockManager) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	return m.locks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		return fn(ctx)
	})
}

func (m leaseRuntimeLockManager) WithSessionListenerAllocationLock(ctx context.Context, fn leaseapp.LockFunc) error {
	return m.locks.WithSessionListenerAllocationLock(ctx, func(ctx context.Context) error {
		return fn(ctx)
	})
}

type randomLeaseIDGenerator struct {
	byteLength int
}

func (g randomLeaseIDGenerator) NewLeaseID() (string, error) {
	return random.Hex(g.byteLength)
}
