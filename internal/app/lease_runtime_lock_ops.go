package app

import (
	"context"
	"errors"
	"strings"
)

func (s *redisLeaseRuntimeLocks) LockAccount(ctx context.Context, accountID string) (leaseRuntimeLock, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, errors.New("lease account_id is required")
	}
	return s.lock(ctx, "account:"+accountID)
}

func (s *redisLeaseRuntimeLocks) LockProviderAccount(ctx context.Context, providerAccountID string) (leaseRuntimeLock, error) {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return nil, errors.New("provider account id is required")
	}
	return s.lock(ctx, "provider-account:"+providerAccountID)
}

func (s *redisLeaseRuntimeLocks) LockSessionListenerAllocation(ctx context.Context) (leaseRuntimeLock, error) {
	return s.lock(ctx, "session-listener-allocation")
}
