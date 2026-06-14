package mihomo

import (
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

const gatewayListenerName = "proxy-runtime-gateway"

func renderGateway(endpoint sourceplane.Endpoint, users []dataplane.ProxyUserRoute, sessions []dataplane.SessionRoute, profiles map[string]string) (mihomoListener, []mihomoGroup, []string, error) {
	host, port, err := splitEndpoint(endpoint.Addr)
	if err != nil {
		return mihomoListener{}, nil, nil, err
	}
	listener := mihomoListener{Name: gatewayListenerName, Type: "mixed", Listen: host, Port: port, UDP: true}
	listener.Users = renderUsers(users, sessions)
	groups := renderUserGroups(users, sessions)
	rules := renderUserRules(users, sessions, profiles)
	return listener, groups, rules, nil
}
