package proxycheck

import (
	"sort"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EdgeCanaryOutcome struct {
	Level        proxygatewayv1.ProxyEdgeAccessRiskLevel
	Score        float64
	Signal       proxygatewayv1.ProxyEdgeAccessRiskSignal
	ErrorMessage string
}

func EdgeBaseFraudCheck(ip string) *proxygatewayv1.ProxyIPFraudCheck {
	return &proxygatewayv1.ProxyIPFraudCheck{
		Ip:        ip,
		RiskLevel: proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_LOW,
		CheckedAt: timestamppb.Now(),
	}
}

func BuildEdgeAccessCheck(
	fraudCheck *proxygatewayv1.ProxyIPFraudCheck,
	expectedCountryCode string,
	outcome EdgeCanaryOutcome,
) *proxygatewayv1.ProxyEdgeAccessCheck {
	check := &proxygatewayv1.ProxyEdgeAccessCheck{
		Ip:           fraudCheck.GetIp(),
		IpFraudCheck: fraudCheck,
		RiskLevel:    proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN,
		RiskScore:    clampEdgeScore(fraudCheck.GetRiskScore()),
		CheckedAt:    fraudCheck.GetCheckedAt(),
		ErrorMessage: outcome.ErrorMessage,
	}
	if outcome.Level == proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNSUPPORTED {
		check.RiskLevel = outcome.Level
		check.RiskSignals = []proxygatewayv1.ProxyEdgeAccessRiskSignal{outcome.Signal}
		return check
	}
	signals := map[proxygatewayv1.ProxyEdgeAccessRiskSignal]struct{}{}
	addSignal := func(signal proxygatewayv1.ProxyEdgeAccessRiskSignal) {
		if signal != proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_UNSPECIFIED {
			signals[signal] = struct{}{}
		}
	}
	for _, signal := range fraudCheck.GetRiskSignals() {
		switch signal {
		case proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_DATACENTER,
			proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_HOSTING:
			addSignal(proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_DATACENTER_NETWORK)
		case proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_PROXY,
			proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_VPN,
			proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_TOR:
			addSignal(proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_ANONYMIZER_DETECTED)
		}
	}
	if fraudCheck.GetRiskLevel() == proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_HIGH ||
		fraudCheck.GetRiskLevel() == proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_CRITICAL {
		addSignal(proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_IP_REPUTATION_RISK)
	}
	if countryMismatch(fraudCheck.GetCountryCode(), expectedCountryCode) {
		addSignal(proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_GEO_MISMATCH)
		check.RiskScore = maxFloat(check.GetRiskScore(), 60)
	}
	addSignal(outcome.Signal)
	check.RiskScore = maxFloat(check.GetRiskScore(), outcome.Score)
	check.RiskLevel = edgeRiskFromScore(check.GetRiskScore())
	check.RiskLevel = maxEdgeRisk(check.GetRiskLevel(), edgeRiskFromIP(fraudCheck.GetRiskLevel()))
	check.RiskLevel = maxEdgeRisk(check.GetRiskLevel(), outcome.Level)
	if outcome.Signal == proxygatewayv1.ProxyEdgeAccessRiskSignal_PROXY_EDGE_ACCESS_RISK_SIGNAL_EDGE_UNAVAILABLE &&
		check.GetRiskLevel() == proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW {
		check.RiskLevel = proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN
	}
	check.RiskSignals = sortedEdgeSignals(signals)
	return check
}

func countryMismatch(actual string, expected string) bool {
	actual = strings.ToUpper(strings.TrimSpace(actual))
	expected = strings.ToUpper(strings.TrimSpace(expected))
	return actual != "" && expected != "" && actual != expected
}

func edgeRiskFromIP(level proxygatewayv1.ProxyIPFraudRiskLevel) proxygatewayv1.ProxyEdgeAccessRiskLevel {
	switch level {
	case proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_CRITICAL:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH
	case proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_HIGH:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH
	case proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_MEDIUM:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_MEDIUM
	case proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_LOW:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW
	default:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN
	}
}

func edgeRiskFromScore(score float64) proxygatewayv1.ProxyEdgeAccessRiskLevel {
	switch {
	case score >= 80:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH
	case score >= 50:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_MEDIUM
	case score > 0:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW
	default:
		return proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN
	}
}

func maxEdgeRisk(
	current proxygatewayv1.ProxyEdgeAccessRiskLevel,
	next proxygatewayv1.ProxyEdgeAccessRiskLevel,
) proxygatewayv1.ProxyEdgeAccessRiskLevel {
	if edgeRiskRank(next) > edgeRiskRank(current) {
		return next
	}
	return current
}

func edgeRiskRank(level proxygatewayv1.ProxyEdgeAccessRiskLevel) int {
	switch level {
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_BLOCK_LIKELY:
		return 70
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_CHALLENGE_LIKELY:
		return 60
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH:
		return 50
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_MEDIUM:
		return 40
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_LOW:
		return 30
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNKNOWN:
		return 20
	case proxygatewayv1.ProxyEdgeAccessRiskLevel_PROXY_EDGE_ACCESS_RISK_LEVEL_UNSUPPORTED:
		return 5
	default:
		return 0
	}
}

func sortedEdgeSignals(values map[proxygatewayv1.ProxyEdgeAccessRiskSignal]struct{}) []proxygatewayv1.ProxyEdgeAccessRiskSignal {
	signals := make([]proxygatewayv1.ProxyEdgeAccessRiskSignal, 0, len(values))
	for signal := range values {
		signals = append(signals, signal)
	}
	sort.Slice(signals, func(i, j int) bool { return signals[i] < signals[j] })
	return signals
}

func clampEdgeScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func maxFloat(left float64, right float64) float64 {
	if right > left {
		return right
	}
	return left
}
