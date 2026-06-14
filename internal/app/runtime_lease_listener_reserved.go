package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func (r *Runtime) listenerReservedLeaseFacts(ctx context.Context) ([]*proxyruntimev1.ProxyDynamicLease, error) {
	active, err := r.store.ListActiveLeaseFacts(ctx, leaseapp.MaxListLimit)
	if err != nil {
		return nil, err
	}
	cleanupPending, err := r.store.CleanupPendingLeaseFacts(ctx)
	if err != nil {
		return nil, err
	}
	return leaseapp.ReservedListenerLeaseFacts(active, cleanupPending), nil
}
