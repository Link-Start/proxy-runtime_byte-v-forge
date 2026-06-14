package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

func (c leaseCoordinator) releaseLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease, err := c.leaseByReleaseRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if leaseapp.HasReleasedStatus(lease) {
		return lease, nil
	}
	accountID := strings.TrimSpace(lease.GetAccountId())
	err = c.deps.locks.WithAccountLock(ctx, accountID, func(ctx context.Context) error {
		current, err := c.deps.store.LeaseFactByID(ctx, lease.GetLeaseId())
		if err != nil && !isStoreNotFound(err) {
			return err
		}
		if current != nil {
			lease = current
		}
		if leaseapp.HasReleasedStatus(lease) {
			return nil
		}
		if !leaseapp.HasActiveStatus(lease) {
			return nil
		}
		return c.retireLeaseRoute(ctx, lease)
	})
	return lease, err
}

func (c leaseCoordinator) leaseByReleaseRequest(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	lookup, err := leaseapp.ParseReleaseRequest(req)
	if err != nil {
		return nil, invalidArgument(err.Error(), err)
	}
	if lookup.LeaseID != "" {
		lease, err := c.deps.store.LeaseFactByID(ctx, lookup.LeaseID)
		if err != nil {
			if isStoreNotFound(err) {
				return nil, invalidArgument("lease_id not found", nil)
			}
			return nil, err
		}
		if err := leaseapp.ValidateReleaseLeaseMatch(lookup, lease); err != nil {
			return nil, invalidArgument(err.Error(), err)
		}
		return lease, nil
	}
	lease, err := c.deps.store.ActiveLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err == nil {
		return lease, nil
	}
	if !isStoreNotFound(err) {
		return nil, err
	}
	lease, err = c.deps.store.LatestLeaseFactByAccount(ctx, lookup.AccountID, lookup.Purpose)
	if err != nil {
		if isStoreNotFound(err) {
			return nil, invalidArgument("active lease not found", nil)
		}
		return nil, err
	}
	return lease, nil
}

func (c leaseCoordinator) retireLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil {
		return nil
	}
	if err := c.deleteLeaseRoute(ctx, lease); err != nil {
		_ = c.saveLeaseReleaseCleanupFailure(ctx, lease, true, false, "lease route cleanup failed")
		return err
	}
	c.clearExitCheckCache()
	if lease.GetAccountId() == playgroundProfileID {
		c.closeMihomoInUserConnections(ctx, []string{playgroundUsername})
	}
	releaseErr := c.releaseLeaseProviderSessionWithLock(ctx, lease)
	if releaseErr != nil {
		c.warn("provider session release failed", leaseapp.LabelAccountID, lease.GetAccountId(), leaseapp.LabelProviderAccountID, lease.GetProviderAccountId())
		if err := c.saveLeaseReleaseCleanupFailure(ctx, lease, false, true, "provider session release failed"); err != nil {
			return err
		}
		return releaseErr
	}
	return c.saveLeaseReleased(ctx, lease)
}

func (c leaseCoordinator) releaseLeaseProviderSessionWithLock(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	providerAccountID := strings.TrimSpace(lease.GetProviderAccountId())
	if providerAccountID == "" {
		return c.releaseLeaseProviderSession(ctx, lease)
	}
	return c.deps.locks.WithProviderAccountLock(ctx, providerAccountID, func(ctx context.Context) error {
		return c.releaseLeaseProviderSession(ctx, lease)
	})
}

func (c leaseCoordinator) deleteLeaseRoute(ctx context.Context, lease *proxyruntimev1.ProxyDynamicLease) error {
	if lease == nil || lease.GetSession() == nil || lease.GetListener() == nil {
		return nil
	}
	listener := listenerFromProto(lease.GetListener())
	route := dataplane.SessionRoute{SessionID: lease.GetSession().GetSessionId(), Listener: localServiceFromListener(listener, c.deps.cfg.LocalProtocol)}
	return c.deps.dataPlane.DeleteSessionRoute(ctx, route)
}
