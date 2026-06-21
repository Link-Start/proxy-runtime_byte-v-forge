package domain

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

func dynamicIPProviderFromProto(in *proxygatewayv1.ProxyDynamicIPProviderSettings) *proxygatewayv1.ProxyDynamicIPProviderSettings {
	if in == nil {
		return &proxygatewayv1.ProxyDynamicIPProviderSettings{}
	}
	out := &proxygatewayv1.ProxyDynamicIPProviderSettings{
		ProviderId:               strings.TrimSpace(in.GetProviderId()),
		DynamicProviderId:        appcore.RuntimeSafeID(in.GetDynamicProviderId()),
		DisplayName:              strings.TrimSpace(in.GetDisplayName()),
		RotatingConcurrencyLimit: kernel.NormalizeDynamicProviderRotatingConcurrencyLimit(in.GetRotatingConcurrencyLimit()),
		StickyConcurrencyLimit:   kernel.NormalizeDynamicProviderStickyConcurrencyLimit(in.GetStickyConcurrencyLimit()),
		Endpoints:                make([]*proxygatewayv1.ProxyDynamicIPEndpointSettings, 0, len(in.GetEndpoints())),
	}
	for _, endpoint := range in.GetEndpoints() {
		out.Endpoints = append(out.Endpoints, dynamicIPEndpointFromProto(endpoint))
	}
	kernel.NormalizeDynamicIPProvider(out)
	return out
}

func dynamicIPEndpointFromProto(in *proxygatewayv1.ProxyDynamicIPEndpointSettings) *proxygatewayv1.ProxyDynamicIPEndpointSettings {
	if in == nil {
		return &proxygatewayv1.ProxyDynamicIPEndpointSettings{}
	}
	out := &proxygatewayv1.ProxyDynamicIPEndpointSettings{
		EndpointUrl: kernel.NormalizeEndpointURL(in.GetEndpointUrl()),
	}
	return out
}

func cloneDynamicIPProvider(in *proxygatewayv1.ProxyDynamicIPProviderSettings) *proxygatewayv1.ProxyDynamicIPProviderSettings {
	return dynamicIPProviderFromProto(in)
}

func cloneDynamicIPEndpoints(in []*proxygatewayv1.ProxyDynamicIPEndpointSettings) []*proxygatewayv1.ProxyDynamicIPEndpointSettings {
	out := make([]*proxygatewayv1.ProxyDynamicIPEndpointSettings, 0, len(in))
	for _, endpoint := range in {
		out = append(out, dynamicIPEndpointFromProto(endpoint))
	}
	return out
}
