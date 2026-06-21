package ipfraud

import (
	"context"
	"net/http"
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
)

type abuseIPDBProvider struct{ httpProvider }

const abuseIPDBEndpoint = "https://api.abuseipdb.com/api/v2/check?ipAddress={ip}&maxAgeInDays=90"

type abuseIPDBPlugin struct{}

func (abuseIPDBPlugin) Kind() proxygatewayv1.ProxyIPFraudProviderKind {
	return proxygatewayv1.ProxyIPFraudProviderKind_PROXY_IP_FRAUD_PROVIDER_KIND_ABUSEIPDB
}
func (abuseIPDBPlugin) ProviderID() string      { return "abuseipdb" }
func (abuseIPDBPlugin) DisplayName() string     { return "AbuseIPDB" }
func (abuseIPDBPlugin) DefaultWeight() uint32   { return 85 }
func (abuseIPDBPlugin) SupportsAnonymous() bool { return false }
func (abuseIPDBPlugin) SupportsAPIKey() bool    { return true }
func (abuseIPDBPlugin) Auth(keys []string, _ bool) AuthConfig {
	return AuthConfig{APIKey: &APIKeyAuthConfig{Keys: append([]string(nil), keys...), Placement: "header", Name: "Key"}}
}
func (abuseIPDBPlugin) New(client *http.Client, cfg ProviderConfig, cooldown time.Duration, clk clock.Clock) provider {
	return &abuseIPDBProvider{httpProvider: newHTTPProvider(client, abuseIPDBEndpoint, cfg.Auth, cooldown, clk)}
}

func (p *abuseIPDBProvider) lookup(ctx context.Context, ip string) (report, error) {
	payload, err := p.lookupJSON(ctx, ip)
	if err != nil {
		return report{}, err
	}
	if _, ok := pathValue(payload, "data"); !ok {
		return report{}, errProviderUnavailable
	}
	return parseAbuseIPDB(payload), nil
}

func parseAbuseIPDB(payload map[string]any) report {
	score, ok := floatValue(payload, "data.abuseConfidenceScore")
	if !ok {
		score = 0
	}
	totalReports, _ := floatValue(payload, "data.totalReports")
	distinctUsers, _ := floatValue(payload, "data.numDistinctUsers")
	usageType := stringValue(payload, "data.usageType")
	flags := riskFlags{
		datacenter: usageIsHosting(usageType),
		hosting:    usageIsHosting(usageType),
		tor:        boolValue(payload, "data.isTor"),
		abuser:     score > 0 || totalReports > 0 || distinctUsers > 0,
		mobile:     strings.Contains(strings.ToLower(usageType), "mobile"),
		crawler:    strings.Contains(strings.ToLower(usageType), "spider"),
	}
	return report{
		networkKind:    networkKindWithFlags(classifyNetworkKind(usageType), flags),
		anonymizerKind: classifyAnonymizerKind(flags.tor, false, false, flags.crawler),
		riskLevel:      riskLevelFromText(""),
		riskScore:      clampScore(score),
		signals:        signalsFromFlags(flags),
		countryCode:    stringValue(payload, "data.countryCode"),
		region:         stringValue(payload, "data.region"),
		city:           stringValue(payload, "data.city"),
		organization:   stringValue(payload, "data.domain"),
		isp:            stringValue(payload, "data.isp"),
	}
}

func usageIsHosting(value string) bool {
	normalized := strings.ToLower(value)
	return strings.Contains(normalized, "data center") ||
		strings.Contains(normalized, "hosting") ||
		strings.Contains(normalized, "transit") ||
		strings.Contains(normalized, "content delivery")
}
