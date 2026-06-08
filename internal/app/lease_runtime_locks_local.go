package app

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type localLeaseRuntimeLocks struct {
	mu    sync.Mutex
	locks map[string]chan struct{}
}

type localLeaseRuntimeLock struct {
	key   string
	owner *localLeaseRuntimeLocks
}

func newLocalLeaseRuntimeLocks() *localLeaseRuntimeLocks {
	return &localLeaseRuntimeLocks{locks: map[string]chan struct{}{}}
}

func (s *localLeaseRuntimeLocks) Close() error { return nil }

func (s *localLeaseRuntimeLocks) WithAccountLock(ctx context.Context, accountID string, fn leaseRuntimeLockFunc) error {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return errors.New("lease account_id is required")
	}
	return s.withLock(ctx, "account:"+accountID, fn)
}

func (s *localLeaseRuntimeLocks) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseRuntimeLockFunc) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return errors.New("provider account id is required")
	}
	return s.withLock(ctx, "provider-account:"+providerAccountID, fn)
}

func (s *localLeaseRuntimeLocks) WithSessionListenerAllocationLock(ctx context.Context, fn leaseRuntimeLockFunc) error {
	return s.withLock(ctx, "session-listener-allocation", fn)
}

func (s *localLeaseRuntimeLocks) withLock(ctx context.Context, key string, fn leaseRuntimeLockFunc) error {
	if fn == nil {
		return errors.New("lease runtime lock function is required")
	}
	lock, err := s.lock(ctx, key)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock(context.Background()) }()
	return fn(ctx)
}

func (s *localLeaseRuntimeLocks) lock(ctx context.Context, key string) (*localLeaseRuntimeLock, error) {
	ch := s.lockChannel(key)
	select {
	case ch <- struct{}{}:
		return &localLeaseRuntimeLock{key: key, owner: s}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *localLeaseRuntimeLocks) lockChannel(key string) chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := s.locks[key]
	if ch == nil {
		ch = make(chan struct{}, 1)
		s.locks[key] = ch
	}
	return ch
}

func (l *localLeaseRuntimeLock) Unlock(context.Context) error {
	if l == nil || l.owner == nil || l.key == "" {
		return nil
	}
	ch := l.owner.lockChannel(l.key)
	select {
	case <-ch:
	default:
	}
	return nil
}
