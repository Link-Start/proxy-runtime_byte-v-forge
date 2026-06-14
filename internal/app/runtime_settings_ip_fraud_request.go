package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

func ipFraudProviderFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPFraudProviderSettings, current map[string][]*commonv1.SecretRef, index int, registry *ipfraud.Registry) (*proxyruntimev1.ProxyIPFraudProviderSettings, error) {
	if in == nil {
		return &proxyruntimev1.ProxyIPFraudProviderSettings{}, nil
	}
	id := strings.TrimSpace(in.GetProviderId())
	if id == "" {
		id = registry.DefaultProviderID(in.GetKind())
	}
	apiKeySecretRefs, err := ipFraudSecretRefsFromRequest(ctx, writer, in, id)
	if err != nil {
		return nil, err
	}
	if len(apiKeySecretRefs) == 0 && !in.GetClearApiKeys() && !in.GetAnonymous() {
		apiKeySecretRefs = current[ipFraudProviderSecretKey(in.GetKind(), id)]
	}
	weight := in.GetWeight()
	if weight == 0 {
		weight = providerDefaultWeight(in.GetKind(), index, registry)
	}
	return &proxyruntimev1.ProxyIPFraudProviderSettings{
		ProviderId:       id,
		DisplayName:      strings.TrimSpace(in.GetDisplayName()),
		Weight:           weight,
		Kind:             in.GetKind(),
		Anonymous:        in.GetAnonymous(),
		ApiKeySecretRefs: apiKeySecretRefs,
	}, nil
}

func ipFraudSecretRefsFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPFraudProviderSettings, providerID string) ([]*commonv1.SecretRef, error) {
	if refs := cleanIPFraudSecretRefs(in.GetApiKeySecretRefs()); len(refs) > 0 {
		return refs, nil
	}
	rawValues := cleanList(in.GetApiKeyValues())
	if len(rawValues) == 0 {
		return nil, nil
	}
	out := make([]*commonv1.SecretRef, 0, len(rawValues))
	for index, raw := range rawValues {
		secretID := secretref.StableID("proxy-runtime-ip-fraud-api-key", fmt.Sprintf("%d", in.GetKind()), providerID, fmt.Sprintf("%d", index))
		saved, err := writeRuntimeSecret(ctx, writer, raw, secretID, ipFraudAPIKeyPurpose)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	return cleanIPFraudSecretRefs(out), nil
}
