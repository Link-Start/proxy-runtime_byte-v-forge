package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
)

type runtimeSettingsProviderViewAdapter struct {
	ipFraudProviders *ipfraud.Registry
	ipGeoProviders   *ipgeo.Registry
}

func newRuntimeSettingsProviderViewAdapter(runtime *Runtime) runtimeSettingsProviderViewAdapter {
	if runtime == nil {
		return runtimeSettingsProviderViewAdapter{}
	}
	return runtimeSettingsProviderViewAdapter{
		ipFraudProviders: runtime.ipFraudProviders,
		ipGeoProviders:   runtime.ipGeoProviders,
	}
}

func (a runtimeSettingsProviderViewAdapter) IPFraudProviderViews() []*proxyruntimev1.ProxyIPFraudProviderDescriptor {
	if a.ipFraudProviders == nil {
		return nil
	}
	return a.ipFraudProviders.ProviderDescriptors()
}

func (a runtimeSettingsProviderViewAdapter) IPGeoProviderViews() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
	if a.ipGeoProviders == nil {
		return nil
	}
	return a.ipGeoProviders.ProviderDescriptors()
}
