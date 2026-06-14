package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type PreparedAcquireInput struct {
	Locks   LockManager
	Request *proxyruntimev1.AcquireProxyLeaseRequest
	Action  AccountLeaseAction
}

func RunPreparedAcquire(ctx context.Context, input PreparedAcquireInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	if err := PrepareAcquireRequest(input.Request); err != nil {
		return nil, err
	}
	return RunAccountLeaseAction(ctx, input.Locks, input.Request.GetAccountId(), input.Action)
}

func IsAcquireRequestError(err error) bool {
	return errors.Is(err, ErrAcquireRequestRequired) ||
		errors.Is(err, ErrAcquireAccountIDRequired)
}
