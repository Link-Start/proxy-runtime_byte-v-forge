package lease

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func SessionRouteFromLease(lease *proxyruntimev1.ProxyDynamicLease, nodes []provider.Node, dialerProxy string, fallbackProtocol string) (SessionRoute, bool) {
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

func ListenerFromProto(listener *proxyruntimev1.EgressListener) Listener {
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

func protocolName(protocol proxyruntimev1.ProxyProtocol) string {
	if protocol == proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5 {
		return "socks5"
	}
	return "http"
}
