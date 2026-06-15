package mihomonative

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func protoFixedProxy(item FixedProxy) *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy{Id: item.ID, Name: item.Name, Type: item.Type, Uri: item.URI}
}

func protoSubscription(item Subscription) *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeSubscription{Id: item.ID, Name: item.Name, Url: item.URL}
}

func FixedProxyFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) FixedProxy {
	if item == nil {
		return FixedProxy{}
	}
	return FixedProxy{ID: item.GetId(), Name: item.GetName(), Type: item.GetType(), URI: item.GetUri()}
}

func SubscriptionFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) Subscription {
	if item == nil {
		return Subscription{}
	}
	return Subscription{ID: item.GetId(), Name: item.GetName(), URL: item.GetUrl()}
}
