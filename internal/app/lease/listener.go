package lease

import (
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

const ListenerModeDynamicSessionLease = "dynamic_ip_session_lease"

var ErrListenerPasswordRequired = errors.New("dynamic lease listener password is not configured")

type Listener struct {
	ID       string
	Addr     string
	Protocol string
	Route    string
	Username string
	Password string
	Labels   map[string]string
}

type ListenerInput struct {
	ID        string
	Addr      string
	Protocol  string
	Route     string
	Username  string
	Password  string
	AccountID string
	LeaseID   string
}

func NewListener(input ListenerInput) (Listener, error) {
	if strings.TrimSpace(input.Password) == "" {
		return Listener{}, ErrListenerPasswordRequired
	}
	return Listener{
		ID:       input.ID,
		Addr:     input.Addr,
		Protocol: input.Protocol,
		Route:    input.Route,
		Username: input.Username,
		Password: input.Password,
		Labels: map[string]string{
			"mode":         ListenerModeDynamicSessionLease,
			LabelAccountID: input.AccountID,
			LabelLeaseID:   input.LeaseID,
		},
	}, nil
}

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

func ReservedListenerLeaseFacts(active []*proxyruntimev1.ProxyDynamicLease, cleanupPending []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.ProxyDynamicLease {
	out := make([]*proxyruntimev1.ProxyDynamicLease, 0, len(active)+len(cleanupPending))
	seen := map[string]struct{}{}
	appendReserved := func(lease *proxyruntimev1.ProxyDynamicLease) {
		if lease.GetListener() == nil {
			return
		}
		leaseID := strings.TrimSpace(lease.GetLeaseId())
		if leaseID != "" {
			if _, exists := seen[leaseID]; exists {
				return
			}
			seen[leaseID] = struct{}{}
		}
		out = append(out, lease)
	}
	for _, lease := range active {
		if HasActiveStatus(lease) {
			appendReserved(lease)
		}
	}
	for _, lease := range cleanupPending {
		if RouteCleanupPending(lease) {
			appendReserved(lease)
		}
	}
	return out
}
