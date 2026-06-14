package app

import (
	"context"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type mihomoNativeResourceReplacement struct {
	ResourceID string
	FixedProxy bool
}

func (s *runtimeSettingsStore) replaceMihomoResourceRefs(ctx context.Context, replacements map[string]mihomoNativeResourceReplacement) (bool, error) {
	if len(replacements) == 0 {
		return false, nil
	}
	return s.mutateRuntimeSettingsIfChanged(ctx, func(settings *runtimeSettingsFile) (bool, error) {
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

func replaceMihomoNodeRef(ref *proxyruntimev1.EgressProfileMihomoNodeRef, replacements map[string]mihomoNativeResourceReplacement) bool {
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

func (s *runtimeSettingsStore) enabledMihomoResourceIDs(ctx context.Context) (map[string]struct{}, error) {
	view, err := s.loadMihomoNativeLocked(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	for _, proxy := range view.GetFixedProxies() {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(proxy), nil)
		addEnabledMihomoResourceID(out, normalized.ID)
		addEnabledMihomoResourceID(out, normalized.Name)
	}
	for _, subscription := range view.GetSubscriptions() {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(subscription), nil)
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
