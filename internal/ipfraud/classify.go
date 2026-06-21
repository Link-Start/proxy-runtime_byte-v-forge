package ipfraud

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func classifyNetworkKind(values ...string) proxygatewayv1.ProxyIPNetworkKind {
	value := strings.ToLower(strings.Join(values, " "))
	switch {
	case strings.Contains(value, "datacenter") || strings.Contains(value, "data center") || strings.Contains(value, "hosting") || strings.Contains(value, "hosted"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_DATACENTER
	case strings.Contains(value, "mobile") || strings.Contains(value, "cellular"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_MOBILE
	case strings.Contains(value, "residential") || strings.Contains(value, "consumer"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_RESIDENTIAL
	case strings.Contains(value, "satellite"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_SATELLITE
	case strings.Contains(value, "broadcast"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_BROADCAST
	case strings.Contains(value, "anycast"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_ANYCAST
	case strings.Contains(value, "business") || strings.Contains(value, "enterprise") || strings.Contains(value, "corporate") ||
		strings.Contains(value, "commercial") || strings.Contains(value, "organization") || strings.Contains(value, "government") ||
		strings.Contains(value, "military") || strings.Contains(value, "university") || strings.Contains(value, "library"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_BUSINESS
	case strings.Contains(value, "isp") || strings.Contains(value, "fixed line") || strings.Contains(value, "cable") || strings.Contains(value, "dsl"):
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_ISP
	default:
		return proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_UNKNOWN
	}
}

func classifyAnonymizerKind(tor, vpn, proxy, crawler bool) proxygatewayv1.ProxyIPAnonymizerKind {
	switch {
	case tor:
		return proxygatewayv1.ProxyIPAnonymizerKind_PROXY_IP_ANONYMIZER_KIND_TOR
	case vpn:
		return proxygatewayv1.ProxyIPAnonymizerKind_PROXY_IP_ANONYMIZER_KIND_VPN
	case proxy:
		return proxygatewayv1.ProxyIPAnonymizerKind_PROXY_IP_ANONYMIZER_KIND_PROXY
	case crawler:
		return proxygatewayv1.ProxyIPAnonymizerKind_PROXY_IP_ANONYMIZER_KIND_CRAWLER
	default:
		return proxygatewayv1.ProxyIPAnonymizerKind_PROXY_IP_ANONYMIZER_KIND_NONE
	}
}

func riskLevelFromText(value string) proxygatewayv1.ProxyIPFraudRiskLevel {
	switch compactLower(value) {
	case "critical", "very_high", "very high", "severe":
		return proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_CRITICAL
	case "high", "risky":
		return proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_HIGH
	case "medium", "moderate":
		return proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_MEDIUM
	case "low", "safe", "clean":
		return proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_LOW
	default:
		return proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_UNKNOWN
	}
}

func signalsFromFlags(flags riskFlags) []proxygatewayv1.ProxyIPFraudSignal {
	signals := []proxygatewayv1.ProxyIPFraudSignal{}
	add := func(enabled bool, signal proxygatewayv1.ProxyIPFraudSignal) {
		if enabled {
			signals = append(signals, signal)
		}
	}
	add(flags.bogon, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_BOGON)
	add(flags.datacenter, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_DATACENTER)
	add(flags.hosting, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_HOSTING)
	add(flags.proxy, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_PROXY)
	add(flags.vpn, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_VPN)
	add(flags.tor, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_TOR)
	add(flags.abuser, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_ABUSER)
	add(flags.crawler, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_CRAWLER)
	add(flags.mobile, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_MOBILE)
	add(flags.satellite, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_SATELLITE)
	add(flags.broadcast, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_BROADCAST)
	add(flags.anycast, proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_ANYCAST)
	return signals
}

func scoreFromFlags(flags riskFlags) float64 {
	score := 0.0
	max := func(value float64) {
		if value > score {
			score = value
		}
	}
	maxIf := func(enabled bool, value float64) {
		if enabled {
			max(value)
		}
	}
	maxIf(flags.bogon, 95)
	maxIf(flags.tor, 90)
	maxIf(flags.abuser, 85)
	maxIf(flags.vpn, 75)
	maxIf(flags.proxy, 70)
	maxIf(flags.crawler, 60)
	maxIf(flags.datacenter || flags.hosting, 55)
	maxIf(flags.broadcast || flags.anycast, 50)
	maxIf(flags.satellite, 45)
	maxIf(flags.mobile, 15)
	return score
}

func networkKindWithFlags(base proxygatewayv1.ProxyIPNetworkKind, flags riskFlags) proxygatewayv1.ProxyIPNetworkKind {
	networkKind := base
	if flags.datacenter || flags.hosting {
		networkKind = chooseNetworkKind(networkKind, proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_DATACENTER)
	}
	if flags.mobile {
		networkKind = chooseNetworkKind(networkKind, proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_MOBILE)
	}
	if flags.satellite {
		networkKind = chooseNetworkKind(networkKind, proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_SATELLITE)
	}
	if flags.broadcast {
		networkKind = chooseNetworkKind(networkKind, proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_BROADCAST)
	}
	if flags.anycast {
		networkKind = chooseNetworkKind(networkKind, proxygatewayv1.ProxyIPNetworkKind_PROXY_IP_NETWORK_KIND_ANYCAST)
	}
	return networkKind
}

func hasNonEmptyString(value map[string]any, paths ...string) bool {
	for _, path := range paths {
		if strings.TrimSpace(stringValue(value, path)) != "" {
			return true
		}
	}
	return false
}

type riskFlags struct {
	bogon      bool
	datacenter bool
	hosting    bool
	proxy      bool
	vpn        bool
	tor        bool
	abuser     bool
	crawler    bool
	mobile     bool
	satellite  bool
	broadcast  bool
	anycast    bool
}
