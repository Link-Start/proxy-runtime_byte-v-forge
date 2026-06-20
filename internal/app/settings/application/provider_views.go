package application

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type IPFraudProviderViews func() []*proxyruntimev1.ProxyIPFraudProviderDescriptor
type IPGeoProviderViews func() []*proxyruntimev1.ProxyIPGeoProviderDescriptor

func (a Application) ListIPFraudProviders(context.Context) (*proxyruntimev1.ListProxyIPFraudProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPFraudProvidersResponse{Providers: a.ipFraudProviders()}, nil
}

func (a Application) ListIPGeoProviders(context.Context) (*proxyruntimev1.ListProxyIPGeoProvidersResponse, error) {
	return &proxyruntimev1.ListProxyIPGeoProvidersResponse{Providers: a.ipGeoProviders()}, nil
}

func (a Application) ipFraudProviders() []*proxyruntimev1.ProxyIPFraudProviderDescriptor {
	if a.ipFraudProviderViews == nil {
		return nil
	}
	return a.ipFraudProviderViews()
}

func (a Application) ipGeoProviders() []*proxyruntimev1.ProxyIPGeoProviderDescriptor {
	if a.ipGeoProviderViews == nil {
		return nil
	}
	return a.ipGeoProviderViews()
}
