package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *Runtime) dynamicProfileLeaseViews() []*proxyruntimev1.ProxyDynamicLease {
	nodes, updatedAt := r.dynamicProfilePoolSnapshot()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Labels["mode"]) != "dynamic_profile" {
			continue
		}
		if !dynamicProfileNodeRequiresLease(node) {
			continue
		}
		if strings.TrimSpace(node.Labels["selected"]) != "true" {
			continue
		}
		out = append(out, dynamicProfileLeaseView(node, updatedAt))
	}
	return out
}

func dynamicProfileNodeRequiresLease(node provider.Node) bool {
	switch strings.TrimSpace(node.Labels["session_mode"]) {
	case proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING.String():
		return false
	default:
		return node.RotationMode != proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST
	}
}

func dynamicProfileLeaseView(node provider.Node, updatedAt time.Time) *proxyruntimev1.ProxyDynamicLease {
	labels := dynamicProfileLeaseLabels(node)
	egress := node.Endpoint()
	egress.Labels = cloneStringMap(labels)
	session := dynamicProfileLeaseSession(node, egress, labels, updatedAt)
	return &proxyruntimev1.ProxyDynamicLease{
		LeaseId:           runtimeSafeID("dynamic-profile-" + node.ID),
		AccountId:         strings.TrimSpace(labels["provider_account_id"]),
		Purpose:           "in-user-profile",
		ProviderAccountId: strings.TrimSpace(labels["provider_account_id"]),
		Status:            proxyruntimev1.ProxyDynamicLeaseStatus_PROXY_DYNAMIC_LEASE_STATUS_ACTIVE,
		Session:           session,
		Egress:            egress,
		AcquiredAt:        timestamppb.New(updatedAt),
		SelectionPlan:     dynamicProfileLeaseSelectionPlan(node, labels, updatedAt),
	}
}

func dynamicProfileLeaseLabels(node provider.Node) map[string]string {
	labels := cloneLabels(node.Labels)
	labels["display_name"] = firstNonEmpty(
		labels["egress_profile_display_name"],
		labels["egress_profile_id"],
		node.ID,
	)
	labels["route_text"] = dynamicProfileRouteText(labels)
	return labels
}

func dynamicProfileLeaseSession(node provider.Node, egress *proxyruntimev1.ProxyEndpoint, labels map[string]string, updatedAt time.Time) *proxyruntimev1.ProxySession {
	return &proxyruntimev1.ProxySession{
		SessionId:  strings.TrimSpace(node.SessionID),
		ProviderId: strings.TrimSpace(node.ProviderID),
		AccountId:  strings.TrimSpace(labels["provider_account_id"]),
		Purpose:    "in-user-profile",
		Policy:     dynamicProfileLeaseSessionPolicy(labels),
		CreatedAt:  timestamppb.New(updatedAt),
		Labels:     cloneStringMap(labels),
		Egress:     egress,
	}
}

func dynamicProfileLeaseSessionPolicy(labels map[string]string) *proxyruntimev1.ProxySessionPolicy {
	policy := dynamicProfileSessionPolicy(&proxyruntimev1.ProxySessionPolicy{
		Mode:      dynamicProfileSessionModeLabel(labels["session_mode"]),
		Region:    labels["region"],
		State:     labels["state"],
		City:      labels["city"],
		Asn:       labels["asn"],
		StickyTtl: dynamicProfileStickyTTLLabel(labels["sticky_ttl"]),
	}, labels["dynamic_ip_endpoint_id"])
	return policy
}

func dynamicProfileSessionModeLabel(value string) proxyruntimev1.ProxySessionMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "proxy_session_mode_rotating", "rotating":
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	default:
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
}

func dynamicProfileStickyTTLLabel(value string) *durationpb.Duration {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || duration <= 0 {
		return nil
	}
	return durationpb.New(duration)
}

func dynamicProfileLeaseSelectionPlan(node provider.Node, labels map[string]string, updatedAt time.Time) *proxyruntimev1.ProxyDynamicIPSelectionPlan {
	return &proxyruntimev1.ProxyDynamicIPSelectionPlan{
		SelectionId: "dynamic-profile-" + shortHash(node.ID),
		SelectedEndpoint: &proxyruntimev1.ProxyDynamicIPEndpointCandidate{
			ProviderAccountId: strings.TrimSpace(labels["provider_account_id"]),
			ProviderId:        strings.TrimSpace(node.ProviderID),
			EndpointId:        strings.TrimSpace(labels["dynamic_ip_endpoint_id"]),
			EndpointUrl:       strings.TrimSpace(labels["dynamic_ip_endpoint_url"]),
			DynamicProviderId: strings.TrimSpace(labels["dynamic_provider_id"]),
			Protocol:          egressProtocol(node),
		},
		SelectionReasons: []string{dynamicProfileRouteText(labels)},
		SelectedAt:       timestamppb.New(updatedAt),
	}
}

func dynamicProfileRouteText(labels map[string]string) string {
	line := firstNonEmpty(labels["line_node_id"], labels["line_resource_id"], "DIRECT")
	exit := firstNonEmpty(labels["dynamic_provider_id"], labels["provider_id"], "dynamic_ip")
	if endpoint := strings.TrimSpace(labels["dynamic_ip_endpoint_url"]); endpoint != "" {
		exit += "/" + endpoint
	}
	return strings.TrimSpace(line) + " -> " + strings.TrimSpace(exit)
}

func egressProtocol(node provider.Node) proxyruntimev1.ProxyProtocol {
	endpoint := node.Endpoint()
	if endpoint == nil {
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
	return endpoint.GetProtocol()
}
