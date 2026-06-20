package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func dynamicIPProviderFromProto(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPProviderSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPProviderSettings{
		ProviderId:               strings.TrimSpace(in.GetProviderId()),
		DynamicProviderId:        appcore.RuntimeSafeID(in.GetDynamicProviderId()),
		DisplayName:              strings.TrimSpace(in.GetDisplayName()),
		RotatingConcurrencyLimit: normalizeDynamicProviderRotatingConcurrencyLimit(in.GetRotatingConcurrencyLimit()),
		StickyConcurrencyLimit:   normalizeDynamicProviderStickyConcurrencyLimit(in.GetStickyConcurrencyLimit()),
		Endpoints:                make([]*proxyruntimev1.ProxyDynamicIPEndpointSettings, 0, len(in.GetEndpoints())),
	}
	for _, endpoint := range in.GetEndpoints() {
		out.Endpoints = append(out.Endpoints, dynamicIPEndpointFromProto(endpoint))
	}
	normalizeDynamicIPProvider(out)
	return out
}

func dynamicIPEndpointFromProto(in *proxyruntimev1.ProxyDynamicIPEndpointSettings) *proxyruntimev1.ProxyDynamicIPEndpointSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPEndpointSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPEndpointSettings{
		EndpointUrl: normalizeEndpointURL(in.GetEndpointUrl()),
	}
	return out
}

func normalizeDynamicIPProvider(provider *proxyruntimev1.ProxyDynamicIPProviderSettings) {
	if provider == nil {
		return
	}
	provider.ProviderId = strings.TrimSpace(provider.GetProviderId())
	provider.DynamicProviderId = appcore.RuntimeSafeID(provider.GetDynamicProviderId())
	if provider.DynamicProviderId == "" {
		provider.DynamicProviderId = appcore.RuntimeSafeID(provider.GetProviderId())
	}
	provider.DisplayName = strings.TrimSpace(provider.GetDisplayName())
	if provider.DisplayName == "" {
		provider.DisplayName = provider.GetDynamicProviderId()
	}
	provider.RotatingConcurrencyLimit = normalizeDynamicProviderRotatingConcurrencyLimit(provider.GetRotatingConcurrencyLimit())
	provider.StickyConcurrencyLimit = normalizeDynamicProviderStickyConcurrencyLimit(provider.GetStickyConcurrencyLimit())
	for index := range provider.Endpoints {
		normalizeDynamicIPEndpoint(provider.Endpoints[index])
	}
}

func normalizeDynamicIPEndpoint(endpoint *proxyruntimev1.ProxyDynamicIPEndpointSettings) {
	if endpoint == nil {
		return
	}
	endpoint.EndpointUrl = normalizeEndpointURL(endpoint.GetEndpointUrl())
}

func cloneDynamicIPProvider(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	return dynamicIPProviderFromProto(in)
}

func cloneDynamicIPEndpoints(in []*proxyruntimev1.ProxyDynamicIPEndpointSettings) []*proxyruntimev1.ProxyDynamicIPEndpointSettings {
	out := make([]*proxyruntimev1.ProxyDynamicIPEndpointSettings, 0, len(in))
	for _, endpoint := range in {
		out = append(out, dynamicIPEndpointFromProto(endpoint))
	}
	return out
}

func dynamicIPProviderID(provider *proxyruntimev1.ProxyDynamicIPProviderSettings) string {
	return appcore.RuntimeSafeID(provider.GetDynamicProviderId())
}
