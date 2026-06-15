package lease

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type ProviderAccountAcquireApplyErrorMapper func(error) error

type ProviderAccountAcquiredRouteApplier struct {
	Applier           AcquiredRouteApplier
	LeaseID           string
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ConcurrencyHolder string
	MapError          ProviderAccountAcquireApplyErrorMapper
}

func (a ProviderAccountAcquiredRouteApplier) Apply(ctx context.Context, acquired ProviderAccountAcquireApplyInput) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := a.Applier.Apply(ctx, AcquiredRouteApplierInput{
		Failure:           acquired.Failure,
		LeaseID:           a.LeaseID,
		Request:           a.Request,
		ProviderClient:    acquired.ProviderClient,
		ProviderAccountID: acquired.ProviderAccountID,
		ConcurrencyHolder: a.ConcurrencyHolder,
		Session:           acquired.Session,
		Nodes:             acquired.Nodes,
		DialerProxy:       acquired.DialerProxy,
		LineLabels:        acquired.LineLabels,
		SelectionPlan:     a.SelectionPlan,
	})
	if err != nil {
		return nil, a.mapError(err)
	}
	return lease, nil
}

func (a ProviderAccountAcquiredRouteApplier) mapError(err error) error {
	if err == nil || a.MapError == nil {
		return err
	}
	return a.MapError(err)
}
