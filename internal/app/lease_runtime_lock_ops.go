package app

import (
	"context"
	"errors"
	"strings"
)

func (s *redisLeaseRuntimeLocks) WithAccountLock(ctx context.Context, accountID string, fn leaseRuntimeLockFunc) error {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return errors.New("lease account_id is required")
	}
	return s.withLock(ctx, "account:"+accountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseRuntimeLockFunc) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return errors.New("provider account id is required")
	}
	return s.withLock(ctx, "provider-account:"+providerAccountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithSessionListenerAllocationLock(ctx context.Context, fn leaseRuntimeLockFunc) error {
	return s.withLock(ctx, "session-listener-allocation", fn)
}
