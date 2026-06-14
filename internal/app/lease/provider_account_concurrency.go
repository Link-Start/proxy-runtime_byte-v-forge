package lease

import (
	"context"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ProviderAccountConcurrencyLimiter interface {
	Acquire(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string, time.Duration) (ProviderAccountConcurrencySlot, error)
	Available(context.Context, string, *proxyruntimev1.ProxySessionPolicy, uint32, string) (bool, error)
	Release(context.Context, string, *proxyruntimev1.ProxySessionPolicy, string) error
}

type ProviderAccountConcurrencySlot interface {
	Release(context.Context) error
}

type NoopProviderAccountConcurrencySlot struct{}

func (NoopProviderAccountConcurrencySlot) Release(context.Context) error {
	return nil
}

func ReleaseConcurrencySlotUnlessKept(ctx context.Context, slot ProviderAccountConcurrencySlot, keep bool) error {
	if keep || slot == nil {
		return nil
	}
	return slot.Release(ctx)
}
