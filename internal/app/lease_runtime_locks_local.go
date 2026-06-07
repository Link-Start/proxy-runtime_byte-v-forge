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

func (s *localLeaseRuntimeLocks) LockAccount(ctx context.Context, accountID string) (leaseRuntimeLock, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, errors.New("lease account_id is required")
	}
	return s.lock(ctx, "account:"+accountID)
}

func (s *localLeaseRuntimeLocks) LockProviderAccount(ctx context.Context, providerAccountID string) (leaseRuntimeLock, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, errors.New("provider account id is required")
	}
	return s.lock(ctx, "provider-account:"+providerAccountID)
}

func (s *localLeaseRuntimeLocks) LockSessionListenerAllocation(ctx context.Context) (leaseRuntimeLock, error) {
	return s.lock(ctx, "session-listener-allocation")
}

func (s *localLeaseRuntimeLocks) lock(ctx context.Context, key string) (leaseRuntimeLock, error) {
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
