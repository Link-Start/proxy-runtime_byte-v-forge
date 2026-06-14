package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var (
	ErrReleaseLeaseIDNotFound = errors.New("lease_id not found")
	ErrActiveLeaseNotFound    = errors.New("active lease not found")
)

type StoreNotFoundFunc func(error) bool

func LookupReleaseLease(ctx context.Context, store OrchestrationStore, req *proxyruntimev1.ReleaseProxyLeaseRequest, isNotFound StoreNotFoundFunc) (*proxyruntimev1.ProxyDynamicLease, error) {
	if store == nil {
		return nil, errors.New("lease store is required")
	}
	lookup, err := ParseReleaseRequest(req)
	if err != nil {
		return nil, err
	}
	if lookup.LeaseID != "" {
		return releaseLeaseByID(ctx, store, lookup, isNotFound)
	}
	return releaseLeaseByAccount(ctx, store, lookup, isNotFound)
}

func IsReleaseLookupRequestError(err error) bool {
	return errors.Is(err, ErrReleaseRequestRequired) ||
		errors.Is(err, ErrReleaseLookupRequired) ||
		errors.Is(err, ErrReleaseAccountIDMismatch) ||
		errors.Is(err, ErrReleasePurposeMismatch) ||
		errors.Is(err, ErrReleaseLeaseIDNotFound) ||
		errors.Is(err, ErrActiveLeaseNotFound)
}

func releaseLeaseByID(ctx context.Context, store OrchestrationStore, lookup ReleaseLookup, isNotFound StoreNotFoundFunc) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := store.LeaseFactByID(ctx, lookup.LeaseID)
	if err != nil {
		if storeNotFound(isNotFound, err) {
			return nil, ErrReleaseLeaseIDNotFound
		}
		return nil, err
	}
	if err := ValidateReleaseLeaseMatch(lookup, lease); err != nil {
		return nil, err
	}
	return lease, nil
}

func releaseLeaseByAccount(ctx context.Context, store OrchestrationStore, lookup ReleaseLookup, isNotFound StoreNotFoundFunc) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := store.ActiveLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err == nil {
		return lease, nil
	}
	if !storeNotFound(isNotFound, err) {
		return nil, err
	}
	lease, err = store.LatestLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err != nil {
		if storeNotFound(isNotFound, err) {
			return nil, ErrActiveLeaseNotFound
		}
		return nil, err
	}
	return lease, nil
}

func storeNotFound(isNotFound StoreNotFoundFunc, err error) bool {
	return isNotFound != nil && isNotFound(err)
}
