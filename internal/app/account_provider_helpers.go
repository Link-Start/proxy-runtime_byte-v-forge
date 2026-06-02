package app

import "github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"

func (r *Runtime) gatewayProtocolForProvider(providerID string, gateway accountproxy.Gateway) string {
	protocol, ok := r.accountProviders.GatewayProtocolForProvider(providerID, gateway)
	if ok {
		return protocol
	}
	return accountproxy.GatewayProtocol(gateway, "socks5")
}
