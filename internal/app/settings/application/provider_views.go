package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type IPFraudProviderViews func() []*proxygatewayv1.ProxyIPFraudProviderDescriptor
type IPGeoProviderViews func() []*proxygatewayv1.ProxyIPGeoProviderDescriptor

func (a Application) ListIPFraudProviders(context.Context) (*proxygatewayv1.ListProxyIPFraudProvidersResponse, error) {
	return &proxygatewayv1.ListProxyIPFraudProvidersResponse{Providers: a.ipFraudProviders()}, nil
}

func (a Application) ListIPGeoProviders(context.Context) (*proxygatewayv1.ListProxyIPGeoProvidersResponse, error) {
	return &proxygatewayv1.ListProxyIPGeoProvidersResponse{Providers: a.ipGeoProviders()}, nil
}

func (a Application) ipFraudProviders() []*proxygatewayv1.ProxyIPFraudProviderDescriptor {
	if a.ipFraudProviderViews == nil {
		return nil
	}
	return a.ipFraudProviderViews()
}

func (a Application) ipGeoProviders() []*proxygatewayv1.ProxyIPGeoProviderDescriptor {
	if a.ipGeoProviderViews == nil {
		return nil
	}
	return a.ipGeoProviderViews()
}
