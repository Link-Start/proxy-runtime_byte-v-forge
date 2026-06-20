package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type PreparedAcquireRunner struct {
	Locks  LockManager
	Action AccountLeaseAction
}

type PreparedAcquireRunnerInput struct {
	Request *proxyruntimev1.AcquireProxyLeaseRequest
}

func (r PreparedAcquireRunner) Run(ctx context.Context, input PreparedAcquireRunnerInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	if err := PrepareAcquireRequest(input.Request); err != nil {
		return nil, err
	}
	return RunAccountLeaseAction(ctx, r.Locks, input.Request.GetAccountId(), r.Action)
}

func IsAcquireRequestError(err error) bool {
	return errors.Is(err, ErrAcquireRequestRequired) ||
		errors.Is(err, ErrAcquireAccountIDRequired)
}
