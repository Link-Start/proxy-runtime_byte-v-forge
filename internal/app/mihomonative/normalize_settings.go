package mihomonative

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func normalizeFixedProxySettings(items []*proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy) []*proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy {
	out := make([]*proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy, 0, len(items))
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

func normalizeSubscriptionSettings(items []*proxygatewayv1.ProxyGatewayMihomoNativeSubscription) []*proxygatewayv1.ProxyGatewayMihomoNativeSubscription {
	out := make([]*proxygatewayv1.ProxyGatewayMihomoNativeSubscription, 0, len(items))
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
