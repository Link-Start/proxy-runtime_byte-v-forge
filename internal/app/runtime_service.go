package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type RuntimeService struct {
	proxyruntimev1.UnimplementedProxyRuntimeServiceServer
	providers runtimeProviderApplication
	leases    runtimeLeaseApplication
	checks    runtimeCheckApplication
	settings  runtimeSettingsApplication
	status    runtimeStatusApplication
}

var _ proxyruntimev1.ProxyRuntimeServiceServer = (*RuntimeService)(nil)

func NewRuntimeService(runtime *Runtime) *RuntimeService {
	return &RuntimeService{
		providers: newRuntimeProviderApplication(runtime),
		leases:    newRuntimeLeaseApplication(runtime),
		checks:    newRuntimeCheckApplication(runtime),
		settings:  newRuntimeSettingsApplication(runtime),
		status:    newRuntimeStatusApplication(runtime),
	}
}

func (r *Runtime) service() *RuntimeService {
	if r.appService != nil {
		return r.appService
	}
	return NewRuntimeService(r)
}
