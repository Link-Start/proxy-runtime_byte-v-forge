package mihomo

import (
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
	"github.com/byte-v-forge/proxy-gateway/internal/sourceplane"
)

const gatewayListenerName = "proxy-gateway-gateway"

type renderedGateway struct {
	listener mihomoListener
	groups   []mihomoGroup
	rules    []string
}

func renderGateway(endpoint sourceplane.Endpoint, users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute, profiles map[string]string) (renderedGateway, error) {
	host, port, err := splitEndpoint(endpoint.Addr)
	if err != nil {
		return renderedGateway{}, err
	}
	listener := mihomoListener{Name: gatewayListenerName, Type: "mixed", Listen: host, Port: port, UDP: true}
	listener.Users = renderUsers(users, sessions)
	return renderedGateway{
		listener: listener,
		groups:   renderUserGroups(users, sessions),
		rules:    renderUserRules(users, sessions, profiles),
	}, nil
}
