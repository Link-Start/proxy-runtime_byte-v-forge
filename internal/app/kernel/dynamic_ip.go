package kernel

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

const (
	DefaultDynamicProviderRotatingConcurrencyLimit uint32 = 10
	DefaultDynamicProviderStickyConcurrencyLimit   uint32 = 2
)

func NormalizeDynamicIPProvider(provider *proxygatewayv1.ProxyDynamicIPProviderSettings) {
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
	provider.RotatingConcurrencyLimit = NormalizeDynamicProviderRotatingConcurrencyLimit(provider.GetRotatingConcurrencyLimit())
	provider.StickyConcurrencyLimit = NormalizeDynamicProviderStickyConcurrencyLimit(provider.GetStickyConcurrencyLimit())
	for index := range provider.Endpoints {
		NormalizeDynamicIPEndpoint(provider.Endpoints[index])
	}
}

func NormalizeDynamicIPEndpoint(endpoint *proxygatewayv1.ProxyDynamicIPEndpointSettings) {
	if endpoint == nil {
		return
	}
	endpoint.EndpointUrl = NormalizeEndpointURL(endpoint.GetEndpointUrl())
}

func NormalizeEndpointURL(value string) string {
	return strings.TrimSpace(value)
}

func EndpointIDFromURL(value string) string {
	value = NormalizeEndpointURL(value)
	if value == "" {
		return ""
	}
	return "endpoint-" + appcore.ShortHash(value)
}

func NormalizeDynamicProviderRotatingConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return DefaultDynamicProviderRotatingConcurrencyLimit
	}
	return value
}

func NormalizeDynamicProviderStickyConcurrencyLimit(value uint32) uint32 {
	if value == 0 {
		return DefaultDynamicProviderStickyConcurrencyLimit
	}
	return value
}

func DynamicIPProviderID(provider *proxygatewayv1.ProxyDynamicIPProviderSettings) string {
	return appcore.RuntimeSafeID(provider.GetDynamicProviderId())
}
