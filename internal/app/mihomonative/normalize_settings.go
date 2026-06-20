package mihomonative

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func normalizeFixedProxySettings(items []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := NormalizeFixedProxy(FixedProxyFromProto(item), nil)
		key := appcore.FirstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, protoFixedProxy(normalized))
	}
	return out
}

func normalizeSubscriptionSettings(items []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := NormalizeSubscription(SubscriptionFromProto(item), nil)
		key := appcore.FirstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, protoSubscription(normalized))
	}
	return out
}
