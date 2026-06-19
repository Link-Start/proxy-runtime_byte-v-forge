package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/clock"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
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
	metrics  *runtimeMetrics
	clock    clock.Clock
}

func (f leaseRegistrySessionProviderFactory) NewSessionProvider(providerCfg accountproxy.Config) (leaseapp.SessionProvider, error) {
	if f.registry == nil {
		return nil, errors.New("provider session factory is required")
	}
	startedAt := time.Now()
	sessionProvider, err := f.registry.NewSessionProvider(providerCfg, f.client, f.clock)
	f.observe(runtimeMetricProviderSessionFactory, startedAt, err)
	if err != nil {
		return nil, err
	}
	return metricLeaseSessionProvider{delegate: sessionProvider, metrics: f.metrics}, nil
}

func (f leaseRegistrySessionProviderFactory) observe(operation string, startedAt time.Time, err error) {
	if f.metrics != nil {
		f.metrics.Observe(operation, startedAt, err)
	}
}

type metricLeaseSessionProvider struct {
	delegate leaseapp.SessionProvider
	metrics  *runtimeMetrics
}

func (p metricLeaseSessionProvider) Name() string {
	if p.delegate == nil {
		return ""
	}
	return p.delegate.Name()
}

func (p metricLeaseSessionProvider) CreateSession(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxySession, error) {
	startedAt := time.Now()
	session, err := p.delegate.CreateSession(ctx, req)
	p.observe(runtimeMetricProviderSessionCreate, startedAt, err)
	return session, err
}

func (p metricLeaseSessionProvider) FetchSession(ctx context.Context, session *proxyruntimev1.ProxySession) ([]provider.Node, error) {
	startedAt := time.Now()
	nodes, err := p.delegate.FetchSession(ctx, session)
	p.observe(runtimeMetricProviderSessionFetch, startedAt, err)
	return nodes, err
}

func (p metricLeaseSessionProvider) ReleaseSession(ctx context.Context, session *proxyruntimev1.ProxySession) error {
	startedAt := time.Now()
	err := p.delegate.ReleaseSession(ctx, session)
	p.observe(runtimeMetricProviderSessionRelease, startedAt, err)
	return err
}

func (p metricLeaseSessionProvider) observe(operation string, startedAt time.Time, err error) {
	if p.metrics != nil {
		p.metrics.Observe(operation, startedAt, err)
	}
}
