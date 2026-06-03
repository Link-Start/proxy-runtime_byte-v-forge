package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/common-lib/secretref"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func settingsFromRequest(ctx context.Context, writer secretref.Writer, req *proxyruntimev1.UpdateProxyRuntimeSettingsRequest, current *runtimeSettingsFile, accountProviders *accountproxy.Registry, ipFraudProviders *ipfraud.Registry) (*runtimeSettingsFile, error) {
	current = normalizeRuntimeSettingsWithProviders(current, ipFraudProviders)
	edgeCanary, err := edgeCanaryFromRequest(ctx, writer, req.GetEdgeCanary(), current.GetEdgeCanary())
	if err != nil {
		return nil, err
	}
	settings := &proxyruntimev1.ProxyRuntimePersistentSettings{
		EdgeCanary:         edgeCanary,
		IpFraudProviders:   make([]*proxyruntimev1.ProxyIPFraudProviderSettings, 0, len(req.GetIpFraudProviders())),
		DynamicIpProviders: make([]*proxyruntimev1.ProxyDynamicIPProviderSettings, 0, len(req.GetDynamicIpProviders())),
		CheckSettings:      checkSettingsFromRequest(req.GetCheckSettings(), current.GetCheckSettings()),
	}
	if edgeCanaryEnabled(settings.GetEdgeCanary()) && strings.TrimSpace(settings.GetEdgeCanary().GetUrl()) == "" {
		return nil, errors.New("edge canary url is required when enabled")
	}
	currentProviders := providerSecrets(current, ipFraudProviders)
	seenProviders := map[string]struct{}{}
	for index, provider := range req.GetIpFraudProviders() {
		item, err := ipFraudProviderFromRequest(ctx, writer, provider, currentProviders, index, ipFraudProviders)
		if err != nil {
			return nil, err
		}
		if err := validateIPFraudProvider(item, index, ipFraudProviders); err != nil {
			return nil, err
		}
		key := providerSecretKey(item.GetKind(), item.GetProviderId())
		if _, exists := seenProviders[key]; exists {
			return nil, fmt.Errorf("ip_fraud_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenProviders[key] = struct{}{}
		settings.IpFraudProviders = append(settings.IpFraudProviders, item)
	}
	seenDynamicProviders := map[string]struct{}{}
	for index, provider := range req.GetDynamicIpProviders() {
		item := dynamicIPProviderFromProto(provider)
		if err := validateDynamicIPProvider(item, index, accountProviders); err != nil {
			return nil, err
		}
		if _, exists := seenDynamicProviders[item.GetProviderId()]; exists {
			return nil, fmt.Errorf("dynamic_ip_providers[%d] duplicates provider %q", index, item.GetProviderId())
		}
		seenDynamicProviders[item.GetProviderId()] = struct{}{}
		settings.DynamicIpProviders = append(settings.DynamicIpProviders, item)
	}
	return normalizeRuntimeSettingsWithProviders(settings, ipFraudProviders), nil
}
