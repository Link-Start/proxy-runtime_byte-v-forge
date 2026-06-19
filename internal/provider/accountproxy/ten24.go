package accountproxy

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/geox"
)

func Ten24Plugin() Plugin {
	return NewDefinitionPlugin(Definition{
		ProviderID:               ProviderTen24,
		DisplayName:              "1024Proxy",
		Default:                  true,
		DefaultProtocol:          "socks5",
		Protocols:                []string{"http", "socks5"},
		UsernameParameterSession: true,
		BuildUsername:            ten24Username,
	})
}

func ten24Username(base string, policy *proxyruntimev1.ProxySessionPolicy, sessionID string) string {
	region := ten24CountryCode(policy.GetRegion())
	if !stickySessionPolicy(policy) {
		return dashUsername(base, "region", region, "st", policy.GetState(), "city", policy.GetCity(), "asn", policy.GetAsn())
	}
	return dashUsername(base, "region", region, "st", policy.GetState(), "city", policy.GetCity(), "asn", policy.GetAsn(), "sid", sessionID, "t", stickyMinutesString(policy))
}

func ten24CountryCode(value string) string {
	country := geox.NormalizeCountryAlpha2(value)
	if country == "GB" {
		return "UK"
	}
	if country != "" {
		return country
	}
	return strings.ToUpper(strings.TrimSpace(value))
}
