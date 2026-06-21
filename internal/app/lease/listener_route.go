package lease

import (
	"context"
	"errors"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

const (
	ListenerRouteDirect   = "direct"
	ListenerRouteProvider = "provider"
)

func NewSessionRoute(sessionID string, listener Listener, nodes []provider.Node, dialerProxy string, fallbackProtocol string) SessionRoute {
	return SessionRoute{
		SessionID:   sessionID,
		Listener:    LocalServiceFromListener(listener, fallbackProtocol),
		Pool:        nodes,
		DialerProxy: dialerProxy,
	}
}

func LocalServiceFromListener(listener Listener, fallbackProtocol string) LocalService {
	return LocalService{
		Name:     listener.ID,
		Addr:     listener.Addr,
		Protocol: ListenerProtocolName(listener, fallbackProtocol),
		Username: listener.Username,
		Password: listener.Password,
		Route:    ListenerRouteName(listener),
	}
}

func ListenerProtocolName(listener Listener, fallback string) string {
	if strings.TrimSpace(listener.Protocol) == "" {
		return fallback
	}
	return listener.Protocol
}

func ListenerRouteName(listener Listener) string {
	switch strings.TrimSpace(listener.Route) {
	case ListenerRouteDirect:
		return ListenerRouteDirect
	default:
		return ListenerRouteProvider
	}
}

var ErrSessionListenerAllocationActionRequired = errors.New("session listener allocation action is required")

type SessionListenerAllocationAction func(context.Context) (*proxygatewayv1.ProxyDynamicLease, error)

func RunSessionListenerAllocation(ctx context.Context, locks LockManager, action SessionListenerAllocationAction) (*proxygatewayv1.ProxyDynamicLease, error) {
	if action == nil {
		return nil, ErrSessionListenerAllocationActionRequired
	}
	var lease *proxygatewayv1.ProxyDynamicLease
	err := WithSessionListenerAllocationLock(ctx, locks, func(ctx context.Context) error {
		var err error
		lease, err = action(ctx)
		return err
	})
	return lease, err
}
