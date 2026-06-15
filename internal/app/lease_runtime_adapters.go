package app

import (
	"context"
	"errors"
	"net/http"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/random"
)

const leaseIDByteLength = 12

type randomLeaseIDGenerator struct {
	byteLength int
}

func (g randomLeaseIDGenerator) NewLeaseID() (string, error) {
	return random.Hex(g.byteLength)
}

type leaseRuntimeLockManager struct {
	locks leaseRuntimeLocks
}

func (m leaseRuntimeLockManager) WithAccountLock(ctx context.Context, accountID string, fn leaseapp.LockFunc) error {
	if m.locks == nil {
		return errors.New("lease runtime locks are required")
	}
	return m.locks.WithAccountLock(ctx, accountID, func(ctx context.Context) error {
		return fn(ctx)
	})
}

func (m leaseRuntimeLockManager) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	if m.locks == nil {
		return errors.New("lease runtime locks are required")
	}
	return m.locks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		return fn(ctx)
	})
}

func (m leaseRuntimeLockManager) WithSessionListenerAllocationLock(ctx context.Context, fn leaseapp.LockFunc) error {
	if m.locks == nil {
		return errors.New("lease runtime locks are required")
	}
	return m.locks.WithSessionListenerAllocationLock(ctx, func(ctx context.Context) error {
		return fn(ctx)
	})
}

type leaseRegistrySessionProviderFactory struct {
	registry *providerregistry.Registry
	client   *http.Client
}

func (f leaseRegistrySessionProviderFactory) NewSessionProvider(providerCfg accountproxy.Config) (leaseapp.SessionProvider, error) {
	if f.registry == nil {
		return nil, errors.New("provider session factory is required")
	}
	return f.registry.NewSessionProvider(providerCfg, f.client)
}
