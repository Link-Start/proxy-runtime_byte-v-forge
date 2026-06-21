package dynamic

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/geox"
)

const countryRegionMatchScore = 600

const sameContinentCountryMatchScore = 450

const sameContinentRegionMatchScore = 350

func hasRequestedRegion(policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) bool {
	return geox.NormalizeCountryAlpha2(policy.GetCountryCode()) != "" || strings.TrimSpace(policy.GetRegion()) != ""
}

func regionScore(regions []string, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) int {
	return regionScoreWithFallback(regions, policy, true)
}

func regionSpecificScore(regions []string, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) int {
	return regionScoreWithFallback(regions, policy, false)
}

func regionScoreWithFallback(regions []string, policy *proxygatewayv1.ProxyDynamicIPSelectionPolicy, includeFallback bool) int {
	regions = cleanRegionCodes(regions)
	if len(regions) == 0 {
		return 0
	}
	country := geox.NormalizeCountryAlpha2(policy.GetCountryCode())
	continent := geox.CountryRegionCode(country)
	requestRegion := strings.ToUpper(strings.TrimSpace(policy.GetRegion()))
	best := 0
	for _, region := range regions {
		regionContinent := geox.NormalizeRegionCode(region)
		endpointCountry := geox.NormalizeCountryAlpha2(region)
		endpointContinent := geox.CountryRegionCode(endpointCountry)
		switch {
		case requestRegion != "" && region == requestRegion:
			best = max(best, 700)
		case country != "" && endpointCountry == country:
			best = max(best, countryRegionMatchScore)
		case continent != "" && endpointContinent == continent:
			best = max(best, sameContinentCountryMatchScore)
		case continent != "" && regionContinent == continent:
			best = max(best, sameContinentRegionMatchScore)
		case includeFallback && (region == "ANY" || region == "GLOBAL" || region == "*"):
			best = max(best, 50)
		}
	}
	return best
}

func regionCodesWithInferred(base []string, values ...string) []string {
	out := cleanRegionCodes(base)
	for _, value := range values {
		for _, country := range geox.CountryCodesInText(value) {
			out = appendRegionCode(out, country)
			if region := geox.CountryRegionCode(country); region != "" {
				out = appendRegionCode(out, region)
			}
		}
	}
	return cleanRegionCodes(out)
}

func appendRegionCode(values []string, value string) []string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func cleanRegionCodes(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
