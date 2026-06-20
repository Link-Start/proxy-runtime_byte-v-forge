package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const dynamicProfileSlotReleaseTimeout = 5 * time.Second

func (r *Runtime) dynamicProfileNodesForSelection(ctx context.Context, client *http.Client, profile *proxyruntimev1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected scoredDynamicIPEndpointCandidate, concurrencyLimit uint32, concurrencyHolder string) []provider.Node {
	profileID := runtimeSafeID(profile.GetProfileId())
	cfg := selection.config
	cfg.Gateways = []accountproxy.Gateway{selected.endpoint}
	providerClient, err := r.accountProviders.NewSessionProvider(cfg, client, r.clock)
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	session := dynamicProfileSession(profileID, selection.accountID, cfg.ProviderID, selected.proto.GetEndpointId(), profile.GetExit().GetDynamicIpPolicy())
	slot, err := leaseapp.AcquireProviderAccountConcurrencySlot(ctx, r.providerConcurrency, selection.account.GetAccountId(), concurrencyLimit, session.GetPolicy(), concurrencyHolder, leaseapp.ConcurrencySlotTTL(session.GetPolicy(), leaseapp.DefaultDynamicIPStickyTTL, providerAccountConcurrencyTTLBuffer))
	if err != nil {
		r.logger.Warn("dynamic profile provider account skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	keepSlot := false
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dynamicProfileSlotReleaseTimeout)
		defer cancel()
		_ = leaseapp.ReleaseConcurrencySlotUnlessKept(releaseCtx, slot, keepSlot)
	}()
	nodes, err := leaseapp.FetchProviderSession(ctx, providerClient, session)
	if err != nil {
		r.logger.Warn("dynamic profile provider session skipped", "account_id", selection.accountID, "provider_id", cfg.ProviderID, "error_type", appcore.ErrorLogType(err))
		return nil
	}
	for index, node := range nodes {
		nodes[index] = dynamicProfileLabelNode(node, index, profile, selection, selected)
	}
	keepSlot = true
	return nodes
}

func dynamicProfileLabelNode(node provider.Node, index int, profile *proxyruntimev1.EgressProfileSettings, selection dynamicProfileEndpointSelection, selected scoredDynamicIPEndpointCandidate) provider.Node {
	profileID := runtimeSafeID(profile.GetProfileId())
	policy := dynamicProfileSessionPolicy(profile.GetExit().GetDynamicIpPolicy(), selected.proto.GetEndpointId())
	node.ID = dynamicProfileNodeID(profileID, selection.accountID, node.SessionID, selected.proto.GetEndpointId(), index)
	node.ProviderID = selection.config.ProviderID
	node.Labels = cloneLabels(node.Labels)
	node.Labels["egress_profile_id"] = profileID
	node.Labels["egress_profile_display_name"] = strings.TrimSpace(profile.GetDisplayName())
	node.Labels["line_kind"] = egressProfileLineKind(profile.GetLine().GetKind())
	node.Labels["line_resource_id"] = strings.TrimSpace(profile.GetLine().GetMihomoNode().GetResourceId())
	node.Labels["line_node_id"] = strings.TrimSpace(profile.GetLine().GetMihomoNode().GetNodeId())
	node.Labels["provider_account_id"] = selection.accountID
	node.Labels["provider_account_display_name"] = strings.TrimSpace(selection.account.GetDisplayName())
	node.Labels["provider_id"] = selection.config.ProviderID
	node.Labels["session_id"] = strings.TrimSpace(node.SessionID)
	node.Labels["dynamic_provider_id"] = selected.proto.GetDynamicProviderId()
	node.Labels["dynamic_provider_ids"] = selected.proto.GetDynamicProviderId()
	node.Labels["dynamic_ip_endpoint_id"] = selected.proto.GetEndpointId()
	node.Labels["session_mode"] = policy.GetMode().String()
	node.Labels["rotation_mode"] = policy.GetRotationMode().String()
	node.Labels["region"] = strings.TrimSpace(policy.GetRegion())
	node.Labels["state"] = strings.TrimSpace(policy.GetState())
	node.Labels["city"] = strings.TrimSpace(policy.GetCity())
	node.Labels["asn"] = strings.TrimSpace(policy.GetAsn())
	node.Labels["sticky_ttl"] = dynamicIPPolicyDurationText(policy)
	node.Labels["mode"] = "dynamic_profile"
	return node
}
