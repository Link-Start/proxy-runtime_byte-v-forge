package mihomonative

import proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

func protoFixedProxy(item FixedProxy) *proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy {
	return &proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy{Id: item.ID, Name: item.Name, Type: item.Type, Uri: item.URI}
}

func protoSubscription(item Subscription) *proxygatewayv1.ProxyGatewayMihomoNativeSubscription {
	return &proxygatewayv1.ProxyGatewayMihomoNativeSubscription{Id: item.ID, Name: item.Name, Url: item.URL}
}

func FixedProxyFromProto(item *proxygatewayv1.ProxyGatewayMihomoNativeFixedProxy) FixedProxy {
	if item == nil {
		return FixedProxy{}
	}
	return FixedProxy{ID: item.GetId(), Name: item.GetName(), Type: item.GetType(), URI: item.GetUri()}
}

func SubscriptionFromProto(item *proxygatewayv1.ProxyGatewayMihomoNativeSubscription) Subscription {
	if item == nil {
		return Subscription{}
	}
	return Subscription{ID: item.GetId(), Name: item.GetName(), URL: item.GetUrl()}
}
