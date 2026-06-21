package app

import (
	"context"
	"sync"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/ipfraud"
	"google.golang.org/protobuf/types/known/timestamppb"

	settingssecret "github.com/byte-v-forge/proxy-gateway/internal/app/settings/adapter/secret"
	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

type ipFraudCheckerCache struct {
	mu        sync.Mutex
	signature string
	checker   ipFraudChecker
}

func (r *Runtime) checkIPFraud(ctx context.Context, ip string, settings *runtimeSettingsFile) (*proxygatewayv1.ProxyIPFraudCheck, error) {
	providers, err := settingssecret.IPFraudProviders(ctx, r.store, settings, r.ipFraudProviders)
	if err != nil {
		return unsupportedIPFraudCheck(ip), nil
	}
	if len(providers) == 0 {
		return unsupportedIPFraudCheck(ip), nil
	}
	service := r.ipFraudChecker(settings, providers)
	return service.Check(ctx, ip)
}

func (r *Runtime) ipFraudChecker(settings *runtimeSettingsFile, providers []ipfraud.ProviderConfig) ipFraudChecker {
	signature := settingsdomain.RuntimeSettingsSignature(settings, r.ipFraudProviders, r.ipGeoProviders)
	return r.fraudChecker.get(signature, func() ipFraudChecker {
		return newIPFraudChecker(r.ipFraudProviders, r.cfg.IPFraud, providers, r.logger, r.clock)
	})
}

func (r *Runtime) resetIPFraudChecker() {
	r.fraudChecker.reset()
}

func (c *ipFraudCheckerCache) get(signature string, create func() ipFraudChecker) ipFraudChecker {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.checker != nil && c.signature == signature {
		return c.checker
	}
	c.checker = create()
	c.signature = signature
	return c.checker
}

func (c *ipFraudCheckerCache) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checker = nil
	c.signature = ""
}

func unsupportedIPFraudCheck(ip string) *proxygatewayv1.ProxyIPFraudCheck {
	return &proxygatewayv1.ProxyIPFraudCheck{
		Ip:        ip,
		RiskLevel: proxygatewayv1.ProxyIPFraudRiskLevel_PROXY_IP_FRAUD_RISK_LEVEL_UNSUPPORTED,
		RiskSignals: []proxygatewayv1.ProxyIPFraudSignal{
			proxygatewayv1.ProxyIPFraudSignal_PROXY_IP_FRAUD_SIGNAL_FRAUD_CHECK_UNSUPPORTED,
		},
		CheckedAt:    timestamppb.Now(),
		ErrorMessage: "IP fraud check is not configured",
	}
}
