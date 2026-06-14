package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func normalizeMihomoNativeFixedProxySettings(items []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) []*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := normalizeMihomoNativeFixedProxy(nativeFixedProxyFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, protoMihomoNativeFixedProxy(normalized))
	}
	return out
}

func normalizeMihomoNativeSubscriptionSettings(items []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) []*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	out := make([]*proxyruntimev1.ProxyRuntimeMihomoNativeSubscription, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		normalized := normalizeMihomoNativeSubscription(nativeSubscriptionFromProto(item), nil)
		key := firstNonEmpty(normalized.ID, normalized.Name)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, protoMihomoNativeSubscription(normalized))
	}
	return out
}
