package lease

import (
	"context"
	"errors"
	"strings"
)

var ErrLockManagerRequired = errors.New("lease lock manager is required")

func WithAccountLock(ctx context.Context, locks LockManager, accountID string, fn LockFunc) error {
	if fn == nil {
		return nil
	}
	if locks == nil {
		return ErrLockManagerRequired
	}
	return locks.WithAccountLock(ctx, strings.TrimSpace(accountID), fn)
}

func WithProviderAccountLock(ctx context.Context, locks LockManager, providerAccountID string, fn LockFunc) error {
	if fn == nil {
		return nil
	}
	providerAccountID = strings.TrimSpace(providerAccountID)
	if providerAccountID == "" {
		return fn(ctx)
	}
	if locks == nil {
		return ErrLockManagerRequired
	}
	return locks.WithProviderAccountLock(ctx, providerAccountID, fn)
}

func WithSessionListenerAllocationLock(ctx context.Context, locks LockManager, fn LockFunc) error {
	if fn == nil {
		return nil
	}
	if locks == nil {
		return ErrLockManagerRequired
	}
	return locks.WithSessionListenerAllocationLock(ctx, fn)
}
