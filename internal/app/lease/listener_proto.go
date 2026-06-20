package lease

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const LabelMode = "mode"

func EgressListenerProto(listener Listener, managed bool, fallbackProtocol string) *proxyruntimev1.EgressListener {
	kind := proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_PROVIDER_ROUTE
	routeID := "default-data-plane"
	if ListenerRouteName(listener) == ListenerRouteDirect {
		kind = proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_DIRECT
		routeID = "direct"
	}
	labels := appcore.CloneStringMap(listener.Labels)
	if listener.Username != "" || listener.Password != "" {
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelProxyUsername] = listener.Username
		labels[LabelProxyPassword] = listener.Password
	}
	if labels[LabelMode] == ListenerModeDynamicSessionLease {
		kind = proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_DYNAMIC_LEASE
		routeID = listener.ID
	}
	return &proxyruntimev1.EgressListener{ListenerId: listener.ID, Kind: kind, ListenAddr: listener.Addr, Protocol: listenerProtocol(listener, fallbackProtocol), RouteId: routeID, Managed: managed, Labels: labels}
}
