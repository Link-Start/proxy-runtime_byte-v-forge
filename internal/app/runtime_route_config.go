package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/config"
)

func (r *Runtime) baseListenerConfigs() []config.EgressListener {
	return defaultListenerConfigs(r.cfg.LocalAddr, r.cfg.LocalProtocol)
}

func (r *Runtime) protoListeners(leases []*proxyruntimev1.ProxyDynamicLease) []*proxyruntimev1.EgressListener {
	return protoListeners(r.baseListenerConfigs(), leases)
}
