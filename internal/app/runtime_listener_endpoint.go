package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) checkIPListener(ctx context.Context, listenerID string) (config.EgressListener, error) {
	listenerID = strings.TrimSpace(listenerID)
	if strings.HasPrefix(listenerID, inUserCheckListenerPrefix) {
		return r.inUserCheckListener(ctx, strings.TrimPrefix(listenerID, inUserCheckListenerPrefix))
	}
	configs := r.baseListenerConfigs()
	leases, err := r.store.ListActiveLeaseFacts(ctx, leaseapp.MaxListLimit)
	if err != nil {
		return config.EgressListener{}, err
	}
	for _, lease := range leases {
		if lease.GetListener() != nil {
			configs = append(configs, listenerFromProto(lease.GetListener()))
		}
	}
	if listenerID != "" {
		for _, listener := range configs {
			if listener.ID == listenerID {
				return listener, nil
			}
		}
		return config.EgressListener{}, fmt.Errorf("listener %q is not configured", listenerID)
	}
	for _, listener := range configs {
		if listenerRoute(listener) == config.ListenerRouteProvider {
			return listener, nil
		}
	}
	if len(configs) == 0 {
		return config.EgressListener{}, errors.New("no egress listener is configured")
	}
	return configs[0], nil
}

const inUserCheckListenerPrefix = "in-user:"

func (r *Runtime) inUserCheckListener(ctx context.Context, username string) (config.EgressListener, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return config.EgressListener{}, errors.New("in-user listener username is required")
	}
	settings, err := r.settings.load(ctx)
	if err != nil {
		return config.EgressListener{}, err
	}
	for _, rule := range settings.GetIngressRules() {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetUsername()) != username {
			continue
		}
		return config.EgressListener{
			ID:       inUserCheckListenerPrefix + runtimeSafeID(username),
			Addr:     r.cfg.LocalAddr,
			Protocol: r.cfg.LocalProtocol,
			Route:    config.ListenerRouteProvider,
			Username: username,
			Password: rule.GetPasswordValue(),
		}, nil
	}
	return config.EgressListener{}, fmt.Errorf("in-user %q is not configured", username)
}

func (r *Runtime) localListenerEndpoint(listener leaseapp.Listener, advertisedHost string) (*proxyruntimev1.ProxyEndpoint, error) {
	return leaseapp.NewListenerEndpoint(listener, advertisedHost, r.cfg.LocalProtocol)
}

func (r *Runtime) sessionAdvertisedHost(advertisedHost string, listener leaseapp.Listener) string {
	if host := strings.TrimSpace(r.cfg.SessionListener.AdvertisedHost); host != "" {
		return host
	}
	bindHost := listenerBindHost(listener.Addr)
	if bindHost != "" && !localOnlyHost(bindHost) {
		return bindHost
	}
	if unspecifiedBindHost(bindHost) {
		if host := firstLocalAdvertisedIP(); host != "" {
			return host
		}
	}
	return strings.TrimSpace(advertisedHost)
}

func listenerBindHost(addr string) string {
	addr = strings.TrimSpace(addr)
	if strings.HasPrefix(addr, ":") {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return strings.Trim(host, "[]")
}

func localOnlyHost(host string) bool {
	host = strings.Trim(strings.TrimSpace(strings.ToLower(host)), "[]")
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip == nil || ip.IsUnspecified() || ip.IsLoopback()
}

func unspecifiedBindHost(host string) bool {
	host = strings.Trim(strings.TrimSpace(strings.ToLower(host)), "[]")
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsUnspecified()
}

func firstLocalAdvertisedIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	ipv6 := ""
	for _, addr := range addrs {
		ip := interfaceIP(addr)
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || !ip.IsGlobalUnicast() {
			continue
		}
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
		if ipv6 == "" {
			ipv6 = ip.String()
		}
	}
	return ipv6
}

func interfaceIP(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP
	case *net.IPAddr:
		return value.IP
	default:
		return nil
	}
}

func advertisedProxyHost(req *http.Request) string {
	if req == nil {
		return ""
	}
	host := strings.TrimSpace(req.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(req.Host)
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		return parsed
	}
	return host
}
