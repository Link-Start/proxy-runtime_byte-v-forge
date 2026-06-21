package persistence

import (
	"context"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
)

func (s *Store) ReplaceMihomoResourceRefs(ctx context.Context, replacements map[string]mihomonative.ResourceReplacement) (bool, error) {
	if len(replacements) == 0 {
		return false, nil
	}
	return s.mutateRuntimeSettingsIfChanged(ctx, func(settings *proxygatewayv1.ProxyGatewayPersistentSettings) (bool, error) {
		changed := false
		for _, profile := range settings.GetEgressProfiles() {
			if replaceMihomoNodeRef(profile.GetLine().GetMihomoNode(), replacements) {
				changed = true
			}
			if replaceMihomoNodeRef(profile.GetExit().GetMihomoNode(), replacements) {
				changed = true
			}
		}
		return changed, nil
	})
}

func replaceMihomoNodeRef(ref *proxygatewayv1.EgressProfileMihomoNodeRef, replacements map[string]mihomonative.ResourceReplacement) bool {
	if ref == nil {
		return false
	}
	current := strings.TrimSpace(ref.GetResourceId())
	replacement, exists := replacements[current]
	if !exists {
		return false
	}
	nextResourceID := strings.TrimSpace(replacement.ResourceID)
	if nextResourceID == "" {
		return false
	}
	oldNodeID := strings.TrimSpace(ref.GetNodeId())
	ref.ResourceId = nextResourceID
	if replacement.FixedProxy {
		ref.NodeId = nextResourceID
		return current != nextResourceID || oldNodeID != nextResourceID
	}
	prefix := current + "/"
	if strings.HasPrefix(oldNodeID, prefix) {
		ref.NodeId = nextResourceID + "/" + strings.TrimPrefix(oldNodeID, prefix)
	}
	return current != nextResourceID
}

func (s *Store) enabledMihomoResourceIDs(ctx context.Context) (map[string]struct{}, error) {
	view, err := s.loadMihomoNativeLocked(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	for _, proxy := range view.GetFixedProxies() {
		normalized := mihomonative.NormalizeFixedProxy(mihomonative.FixedProxyFromProto(proxy), nil)
		addEnabledMihomoResourceID(out, normalized.ID)
		addEnabledMihomoResourceID(out, normalized.Name)
	}
	for _, subscription := range view.GetSubscriptions() {
		normalized := mihomonative.NormalizeSubscription(mihomonative.SubscriptionFromProto(subscription), nil)
		addEnabledMihomoResourceID(out, normalized.ID)
		addEnabledMihomoResourceID(out, normalized.Name)
	}
	return out, nil
}

func addEnabledMihomoResourceID(resources map[string]struct{}, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		resources[value] = struct{}{}
	}
}
