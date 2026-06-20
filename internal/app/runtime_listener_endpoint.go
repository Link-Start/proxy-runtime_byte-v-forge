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

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
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
	settings, err := r.settings.Load(ctx)
	if err != nil {
		return config.EgressListener{}, err
	}
	for _, rule := range settings.GetIngressRules() {
		if !rule.GetEnabled() || strings.TrimSpace(rule.GetUsername()) != username {
			continue
		}
		return config.EgressListener{
			ID:       inUserCheckListenerPrefix + appcore.RuntimeSafeID(username),
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
	return leaseapp.AdvertisedHost(r.cfg.SessionListener.AdvertisedHost, advertisedHost, listener)
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
