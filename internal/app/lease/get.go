package lease

import (
	"context"
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrLeaseIDRequired = errors.New("lease_id is required")

func (a *Application) Get(ctx context.Context, leaseID string) (*proxyruntimev1.ProxyDynamicLease, error) {
	leaseID = strings.TrimSpace(leaseID)
	if leaseID == "" {
		return nil, ErrLeaseIDRequired
	}
	if a == nil || a.repository == nil {
		return nil, errors.New("lease repository is required")
	}
	return a.repository.LeaseFactByID(ctx, leaseID)
}
