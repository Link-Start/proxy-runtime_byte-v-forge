package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/kernel"
)

func dynamicIPProviderFromProto(in *proxyruntimev1.ProxyDynamicIPProviderSettings) *proxyruntimev1.ProxyDynamicIPProviderSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPProviderSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPProviderSettings{
		ProviderId:               strings.TrimSpace(in.GetProviderId()),
		DynamicProviderId:        appcore.RuntimeSafeID(in.GetDynamicProviderId()),
		DisplayName:              strings.TrimSpace(in.GetDisplayName()),
		RotatingConcurrencyLimit: kernel.NormalizeDynamicProviderRotatingConcurrencyLimit(in.GetRotatingConcurrencyLimit()),
		StickyConcurrencyLimit:   kernel.NormalizeDynamicProviderStickyConcurrencyLimit(in.GetStickyConcurrencyLimit()),
		Endpoints:                make([]*proxyruntimev1.ProxyDynamicIPEndpointSettings, 0, len(in.GetEndpoints())),
	}
	for _, endpoint := range in.GetEndpoints() {
		out.Endpoints = append(out.Endpoints, dynamicIPEndpointFromProto(endpoint))
	}
	kernel.NormalizeDynamicIPProvider(out)
	return out
}

func dynamicIPEndpointFromProto(in *proxyruntimev1.ProxyDynamicIPEndpointSettings) *proxyruntimev1.ProxyDynamicIPEndpointSettings {
	if in == nil {
		return &proxyruntimev1.ProxyDynamicIPEndpointSettings{}
	}
	out := &proxyruntimev1.ProxyDynamicIPEndpointSettings{
		EndpointUrl: kernel.NormalizeEndpointURL(in.GetEndpointUrl()),
	}
	return out
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
