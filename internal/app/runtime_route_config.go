package app

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/config"
)

func (r *Runtime) baseListenerConfigs() []config.EgressListener {
	return defaultListenerConfigs(r.cfg.LocalAddr, r.cfg.LocalProtocol)
}

func (r *Runtime) protoListeners(leases []*proxygatewayv1.ProxyDynamicLease) []*proxygatewayv1.EgressListener {
	return protoListeners(r.baseListenerConfigs(), leases)
}
