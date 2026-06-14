package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func RefreshReleaseLease(ctx context.Context, store OrchestrationStore, lease *proxyruntimev1.ProxyDynamicLease, isNotFound StoreNotFoundFunc) (*proxyruntimev1.ProxyDynamicLease, error) {
	if store == nil || !HasLeaseID(lease) {
		return lease, nil
	}
	current, err := store.LeaseFactByID(ctx, lease.GetLeaseId())
	if err != nil {
		if storeNotFound(isNotFound, err) {
			return lease, nil
		}
		return nil, err
	}
	return current, nil
}

func ReleaseNeedsRouteRetire(lease *proxyruntimev1.ProxyDynamicLease) bool {
	return HasActiveStatus(lease) && !HasReleasedStatus(lease)
}
