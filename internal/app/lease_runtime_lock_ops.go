package app

import (
	"context"
	"errors"
	"strings"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (s *redisLeaseRuntimeLocks) WithAccountLock(ctx context.Context, accountID string, fn leaseapp.LockFunc) error {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return errors.New("lease account_id is required")
	}
	return s.withLock(ctx, "account:"+accountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithProviderAccountLock(ctx context.Context, providerAccountID string, fn leaseapp.LockFunc) error {
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return errors.New("provider account id is required")
	}
	return s.withLock(ctx, "provider-account:"+providerAccountID, fn)
}

func (s *redisLeaseRuntimeLocks) WithSessionListenerAllocationLock(ctx context.Context, fn leaseapp.LockFunc) error {
	return s.withLock(ctx, "session-listener-allocation", fn)
}
