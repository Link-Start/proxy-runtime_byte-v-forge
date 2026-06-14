package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func defaultListenerConfigs(localAddr string, localProtocol string) []config.EgressListener {
	return []config.EgressListener{{ID: "dynamic-egress", Addr: localAddr, Protocol: localProtocol, Route: config.ListenerRouteProvider}}
}

func protoListeners(configs []config.EgressListener, leases []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.EgressListener {
	out := make([]*proxyruntimev1.EgressListener, 0, len(configs)+len(leases))
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

func protoListener(listener config.EgressListener, managed bool) *proxyruntimev1.EgressListener {
	route := listenerRoute(listener)
	kind := proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_PROVIDER_ROUTE
	routeID := "default-data-plane"
	if route == config.ListenerRouteDirect {
		kind = proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_DIRECT
		routeID = "direct"
	}
	labels := cloneLabels(listener.Labels)
	if listener.Username != "" || listener.Password != "" {
		if labels == nil {
			labels = map[string]string{}
		}
		labels["proxy_username"] = listener.Username
		labels["proxy_password"] = listener.Password
	}
	if labels["mode"] == "dynamic_ip_session_lease" {
		kind = proxyruntimev1.EgressListenerKind_EGRESS_LISTENER_KIND_DYNAMIC_LEASE
		routeID = listener.ID
	}
	return &proxyruntimev1.EgressListener{ListenerId: listener.ID, Kind: kind, ListenAddr: listener.Addr, Protocol: protocolFromName(listenerProtocol(listener, "http")), RouteId: routeID, Managed: managed, Labels: labels}
}

func localServiceFromListener(listener config.EgressListener, fallback string) leaseapp.LocalService {
	return leaseapp.LocalService{Name: listener.ID, Addr: listener.Addr, Protocol: listenerProtocol(listener, fallback), Username: listener.Username, Password: listener.Password, Route: listenerRoute(listener)}
}

func listenerFromProto(listener *proxyruntimev1.EgressListener) config.EgressListener {
	if listener == nil {
		return config.EgressListener{}
	}
	labels := listener.GetLabels()
	return config.EgressListener{ID: listener.GetListenerId(), Addr: listener.GetListenAddr(), Protocol: protocolName(listener.GetProtocol()), Route: config.ListenerRouteProvider, Username: labels["proxy_username"], Password: labels["proxy_password"], Labels: labels}
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

func protoLeaseListener(listener leaseapp.Listener, managed bool) *proxyruntimev1.EgressListener {
	return protoListener(configListenerFromLease(listener), managed)
}

func localServiceFromLeaseListener(listener leaseapp.Listener, fallback string) leaseapp.LocalService {
	return leaseapp.LocalService{Name: listener.ID, Addr: listener.Addr, Protocol: leaseListenerProtocol(listener, fallback), Username: listener.Username, Password: listener.Password, Route: leaseListenerRoute(listener)}
}

func configListenerFromLease(listener leaseapp.Listener) config.EgressListener {
	return config.EgressListener{ID: listener.ID, Addr: listener.Addr, Protocol: listener.Protocol, Route: listener.Route, Username: listener.Username, Password: listener.Password, Labels: listener.Labels}
}

func leaseListenerProtocol(listener leaseapp.Listener, fallback string) string {
	if strings.TrimSpace(listener.Protocol) == "" {
		return fallback
	}
	return listener.Protocol
}

func leaseListenerRoute(listener leaseapp.Listener) string {
	switch strings.TrimSpace(listener.Route) {
	case config.ListenerRouteDirect:
		return config.ListenerRouteDirect
	default:
		return config.ListenerRouteProvider
	}
}
