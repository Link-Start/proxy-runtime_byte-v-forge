package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func protoMihomoNativeFixedProxy(item mihomoNativeFixedProxy) *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy{Id: item.ID, Name: item.Name, Type: item.Type, Uri: item.URI}
}

func protoMihomoNativeSubscription(item mihomoNativeSubscription) *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription {
	return &proxyruntimev1.ProxyRuntimeMihomoNativeSubscription{Id: item.ID, Name: item.Name, Url: item.URL}
}

func nativeFixedProxyFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeFixedProxy) mihomoNativeFixedProxy {
	if item == nil {
		return mihomoNativeFixedProxy{}
	}
	return mihomoNativeFixedProxy{ID: item.GetId(), Name: item.GetName(), Type: item.GetType(), URI: item.GetUri()}
}

func nativeSubscriptionFromProto(item *proxyruntimev1.ProxyRuntimeMihomoNativeSubscription) mihomoNativeSubscription {
	if item == nil {
		return mihomoNativeSubscription{}
	}
	return mihomoNativeSubscription{ID: item.GetId(), Name: item.GetName(), URL: item.GetUrl()}
}
