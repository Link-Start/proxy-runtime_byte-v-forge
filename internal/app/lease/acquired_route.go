package lease

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type AcquiredSessionRouteInput struct {
	Session       *proxyruntimev1.ProxySession
	Egress        *proxyruntimev1.ProxyEndpoint
	Listener      Listener
	Nodes         []provider.Node
	DialerProxy   string
	LocalProtocol string
}

func NewAcquiredSessionRoute(input AcquiredSessionRouteInput) SessionRoute {
	if input.Session != nil {
		input.Session.Egress = input.Egress
	}
	return NewSessionRoute(input.Session.GetSessionId(), input.Listener, input.Nodes, input.DialerProxy, input.LocalProtocol)
}
