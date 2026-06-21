package lease

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

func SessionRouteFromLease(lease *proxygatewayv1.ProxyDynamicLease, nodes []provider.Node, dialerProxy string, fallbackProtocol string) (SessionRoute, bool) {
	if lease == nil || lease.GetSession() == nil || lease.GetListener() == nil {
		return SessionRoute{}, false
	}
	return SessionRoute{
		SessionID:   lease.GetSession().GetSessionId(),
		Listener:    LocalServiceFromListener(ListenerFromProto(lease.GetListener()), fallbackProtocol),
		Pool:        nodes,
		DialerProxy: dialerProxy,
	}, true
}

func ListenerFromProto(listener *proxygatewayv1.EgressListener) Listener {
	if listener == nil {
		return Listener{}
	}
	labels := listener.GetLabels()
	return Listener{
		ID:       listener.GetListenerId(),
		Addr:     listener.GetListenAddr(),
		Protocol: protocolName(listener.GetProtocol()),
		Route:    ListenerRouteProvider,
		Username: labels[LabelProxyUsername],
		Password: labels[LabelProxyPassword],
		Labels:   labels,
	}
}

func protocolName(protocol proxygatewayv1.ProxyProtocol) string {
	if protocol == proxygatewayv1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5 {
		return "socks5"
	}
	return "http"
}
