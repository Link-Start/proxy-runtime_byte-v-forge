package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

type leaseDynamicIPSelectionAdapter struct {
	selector *dynamicIPSelector
}

func (a leaseDynamicIPSelectionAdapter) Select(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (leaseapp.DynamicIPSelection, error) {
	if a.selector == nil {
		return leaseapp.DynamicIPSelection{}, appcore.InternalError("dynamic IP selector is not configured", nil)
	}
	return a.selector.selectDynamicIPEndpoint(ctx, req)
}

func mapDynamicIPSelectionError(err error) error {
	return appcore.FailedPrecondition("no dynamic IP endpoint candidate", err)
}
