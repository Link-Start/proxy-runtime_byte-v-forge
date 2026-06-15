package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

type leaseDynamicIPSelectionAdapter struct {
	selector *dynamicIPSelector
}

func (a leaseDynamicIPSelectionAdapter) Select(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (leaseapp.DynamicIPSelection, error) {
	if a.selector == nil {
		return leaseapp.DynamicIPSelection{}, internalError("dynamic IP selector is not configured", nil)
	}
	return a.selector.selectDynamicIPEndpoint(ctx, req)
}

func mapDynamicIPSelectionError(err error) error {
	return failedPrecondition("no dynamic IP endpoint candidate", err)
}
