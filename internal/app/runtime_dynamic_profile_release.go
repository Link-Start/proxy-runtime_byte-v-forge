package app

import (
	"context"
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errDynamicProfileLeaseNotFound = errors.New("dynamic profile lease not found")

func dynamicProfileLeaseReleaseRequest(req *proxyruntimev1.ReleaseProxyLeaseRequest) bool {
	return req != nil && strings.HasPrefix(strings.TrimSpace(req.GetLeaseId()), "dynamic-profile-")
}

func (r *Runtime) releaseDynamicProfileLease(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (*proxyruntimev1.ProxyDynamicLease, error) {
	node, profile, lease, err := r.dynamicProfileReleaseTarget(ctx, req)
	if err != nil {
		return nil, err
	}
	accountID := strings.TrimSpace(lease.GetProviderAccountId())
	lock, err := r.leaseLocks.LockAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Unlock(ctx) }()
	sessionKey := dynamicProfileSessionStateKeyForNode(profile, node)
	if r.store != nil {
		if err := r.store.MarkDynamicProfileSessionReleased(ctx, sessionKey); err != nil {
			return nil, err
		}
	}
	lease.Status = proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_RELEASED
	lease.ExpiresAt = timestamppb.Now()
	if lease.GetSession() != nil {
		lease.Session.Policy = dynamicProfileSessionPolicy(profile.GetExit().GetDynamicIpPolicy(), node.Labels["dynamic_ip_endpoint_id"])
	}
	releaseErr := r.leaseCoordinator.releaseLeaseProviderSession(ctx, lease)
	reconcileErr := r.runReconcile(ctx)
	if releaseErr != nil {
		return nil, releaseErr
	}
	if reconcileErr != nil {
		return nil, reconcileErr
	}
	if err := r.releaseLeaseConcurrencySlot(ctx, lease); err != nil {
		r.logger.Warn("release dynamic profile concurrency slot failed", "provider_account_id", lease.GetProviderAccountId())
	}
	return lease, nil
}

func (r *Runtime) dynamicProfileReleaseTarget(ctx context.Context, req *proxyruntimev1.ReleaseProxyLeaseRequest) (provider.Node, *proxyruntimev1.EgressProfileSettings, *proxyruntimev1.ProxyDynamicLease, error) {
	if req == nil {
		return provider.Node{}, nil, nil, errors.New("release request is required")
	}
	nodes, updatedAt := r.dynamicProfilePoolSnapshot()
	if len(nodes) == 0 {
		return provider.Node{}, nil, nil, errDynamicProfileLeaseNotFound
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		return provider.Node{}, nil, nil, err
	}
	profiles := dynamicProfileReleaseProfiles(settings)
	for _, node := range nodes {
		if strings.TrimSpace(node.Labels["mode"]) != "dynamic_profile" || strings.TrimSpace(node.Labels["selected"]) != "true" {
			continue
		}
		if !dynamicProfileNodeRequiresLease(node) {
			continue
		}
		lease := dynamicProfileLeaseView(node, updatedAt)
		if !dynamicProfileLeaseMatchesReleaseRequest(lease, req) {
			continue
		}
		profile := profiles[runtimeSafeID(node.Labels["egress_profile_id"])]
		if profile == nil {
			return provider.Node{}, nil, nil, errors.New("dynamic profile setting not found")
		}
		return node, profile, lease, nil
	}
	return provider.Node{}, nil, nil, errDynamicProfileLeaseNotFound
}

func dynamicProfileReleaseProfiles(settings *runtimeSettingsFile) map[string]*proxyruntimev1.EgressProfileSettings {
	out := map[string]*proxyruntimev1.EgressProfileSettings{}
	for _, profile := range normalizeRuntimeSettings(settings).GetEgressProfiles() {
		out[runtimeSafeID(profile.GetProfileId())] = profile
	}
	return out
}

func dynamicProfileLeaseMatchesReleaseRequest(lease *proxyruntimev1.ProxyDynamicLease, req *proxyruntimev1.ReleaseProxyLeaseRequest) bool {
	leaseID := strings.TrimSpace(req.GetLeaseId())
	if leaseID != "" && leaseID != lease.GetLeaseId() {
		return false
	}
	accountID := strings.TrimSpace(req.GetAccountId())
	if accountID != "" && accountID != lease.GetAccountId() && accountID != lease.GetProviderAccountId() {
		return false
	}
	purpose := strings.TrimSpace(req.GetPurpose())
	if purpose != "" && purpose != lease.GetPurpose() {
		return false
	}
	return leaseID != "" || accountID != ""
}

func dynamicProfileSessionStateKeyForNode(profile *proxyruntimev1.EgressProfileSettings, node provider.Node) string {
	if profile == nil {
		return ""
	}
	return dynamicProfileSessionStateKey(
		profile.GetProfileId(),
		node.Labels["provider_account_id"],
		firstNonEmpty(node.Labels["provider_id"], node.ProviderID),
		node.Labels["dynamic_ip_endpoint_id"],
		profile.GetExit().GetDynamicIpPolicy(),
	)
}
