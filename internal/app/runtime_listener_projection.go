package app

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/byte-v-forge/proxy-gateway/internal/config"
)

func defaultListenerConfigs(localAddr string, localProtocol string) []config.EgressListener {
	return []config.EgressListener{{ID: "dynamic-egress", Addr: localAddr, Protocol: localProtocol, Route: config.ListenerRouteProvider}}
}

func protoListeners(configs []config.EgressListener, leases []*proxygatewayv1.ProxyDynamicLease) []*proxygatewayv1.EgressListener {
	out := make([]*proxygatewayv1.EgressListener, 0, len(configs)+len(leases))
	for _, listener := range configs {
		out = append(out, protoListener(listener, true))
	}
	for _, lease := range leases {
		if lease.GetListener() != nil {
			out = append(out, lease.GetListener())
		}
	}
	return out
}

func protoListener(listener config.EgressListener, managed bool) *proxygatewayv1.EgressListener {
	route := listenerRoute(listener)
	kind := proxygatewayv1.EgressListenerKind_EGRESS_LISTENER_KIND_PROVIDER_ROUTE
	routeID := "default-data-plane"
	if route == config.ListenerRouteDirect {
		kind = proxygatewayv1.EgressListenerKind_EGRESS_LISTENER_KIND_DIRECT
		routeID = "direct"
	}
	labels := cloneLabels(listener.Labels)
	if listener.Username != "" || listener.Password != "" {
		if labels == nil {
			labels = map[string]string{}
		}
		labels[leaseapp.LabelProxyUsername] = listener.Username
		labels[leaseapp.LabelProxyPassword] = listener.Password
	}
	if labels["mode"] == leaseapp.ListenerModeDynamicSessionLease {
		kind = proxygatewayv1.EgressListenerKind_EGRESS_LISTENER_KIND_DYNAMIC_LEASE
		routeID = listener.ID
	}
	return &proxygatewayv1.EgressListener{ListenerId: listener.ID, Kind: kind, ListenAddr: listener.Addr, Protocol: protocolFromName(listenerProtocol(listener, "http")), RouteId: routeID, Managed: managed, Labels: labels}
}

func localServiceFromListener(listener config.EgressListener, fallback string) leaseapp.LocalService {
	return leaseapp.LocalService{Name: listener.ID, Addr: listener.Addr, Protocol: listenerProtocol(listener, fallback), Username: listener.Username, Password: listener.Password, Route: listenerRoute(listener)}
}

func listenerFromProto(listener *proxygatewayv1.EgressListener) config.EgressListener {
	if listener == nil {
		return config.EgressListener{}
	}
	labels := listener.GetLabels()
	return config.EgressListener{ID: listener.GetListenerId(), Addr: listener.GetListenAddr(), Protocol: protocolName(listener.GetProtocol()), Route: config.ListenerRouteProvider, Username: labels[leaseapp.LabelProxyUsername], Password: labels[leaseapp.LabelProxyPassword], Labels: labels}
}

func listenerProtocol(listener config.EgressListener, fallback string) string {
	if strings.TrimSpace(listener.Protocol) == "" {
		return fallback
	}
	return listener.Protocol
}

func listenerRoute(listener config.EgressListener) string {
	switch strings.TrimSpace(listener.Route) {
	case config.ListenerRouteDirect:
		return config.ListenerRouteDirect
	default:
		return config.ListenerRouteProvider
	}
}
