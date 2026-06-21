package ipgeo

import (
	"net/http"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type ipinfoPlugin struct{}
type ip2LocationPlugin struct{}
type ipAPIComPlugin struct{}

func (ipinfoPlugin) Kind() proxygatewayv1.ProxyIPGeoProviderKind {
	return proxygatewayv1.ProxyIPGeoProviderKind_PROXY_IP_GEO_PROVIDER_KIND_IPINFO
}
func (ipinfoPlugin) ProviderID() string      { return "ipinfo" }
func (ipinfoPlugin) DisplayName() string     { return "IPinfo" }
func (ipinfoPlugin) DefaultWeight() uint32   { return 80 }
func (ipinfoPlugin) SupportsAnonymous() bool { return false }
func (ipinfoPlugin) SupportsAPIKey() bool    { return true }
func (ipinfoPlugin) Auth(keys []string, _ bool) AuthConfig {
	return AuthConfig{APIKey: &APIKeyAuthConfig{Keys: append([]string(nil), keys...), Placement: "query", Name: "token"}}
}
func (ipinfoPlugin) New(client *http.Client, cfg ProviderConfig) provider {
	return newHTTPProvider(client, "https://ipinfo.io/{ip}/json", cfg.Auth)
}

func (ip2LocationPlugin) Kind() proxygatewayv1.ProxyIPGeoProviderKind {
	return proxygatewayv1.ProxyIPGeoProviderKind_PROXY_IP_GEO_PROVIDER_KIND_IP2LOCATION
}
func (ip2LocationPlugin) ProviderID() string      { return "ip2location" }
func (ip2LocationPlugin) DisplayName() string     { return "IP2Location.io" }
func (ip2LocationPlugin) DefaultWeight() uint32   { return 70 }
func (ip2LocationPlugin) SupportsAnonymous() bool { return true }
func (ip2LocationPlugin) SupportsAPIKey() bool    { return true }
func (ip2LocationPlugin) Auth(keys []string, anonymous bool) AuthConfig {
	if anonymous {
		return AuthConfig{Anonymous: &AnonymousAuthConfig{}}
	}
	return AuthConfig{APIKey: &APIKeyAuthConfig{Keys: append([]string(nil), keys...), Placement: "query", Name: "key"}}
}
func (ip2LocationPlugin) New(client *http.Client, cfg ProviderConfig) provider {
	return newHTTPProvider(client, "https://api.ip2location.io/?ip={ip}", cfg.Auth)
}

func (ipAPIComPlugin) Kind() proxygatewayv1.ProxyIPGeoProviderKind {
	return proxygatewayv1.ProxyIPGeoProviderKind_PROXY_IP_GEO_PROVIDER_KIND_IP_API_COM
}
func (ipAPIComPlugin) ProviderID() string      { return "ip-api-com" }
func (ipAPIComPlugin) DisplayName() string     { return "IP-API.com" }
func (ipAPIComPlugin) DefaultWeight() uint32   { return 40 }
func (ipAPIComPlugin) SupportsAnonymous() bool { return true }
func (ipAPIComPlugin) SupportsAPIKey() bool    { return false }
func (ipAPIComPlugin) Auth(_ []string, _ bool) AuthConfig {
	return AuthConfig{Anonymous: &AnonymousAuthConfig{}}
}
func (ipAPIComPlugin) New(client *http.Client, cfg ProviderConfig) provider {
	return newHTTPProvider(client, "http://ip-api.com/json/{ip}?fields=status,message,query,countryCode,regionName,city", cfg.Auth)
}
