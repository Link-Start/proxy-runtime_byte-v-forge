package app

import (
	"context"
	"fmt"
	"strings"

	commonv1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/common/v1"
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"

	"github.com/byte-v-forge/proxy-runtime/internal/app/settingscore"
)

func ipGeoProviderFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPGeoProviderSettings, current map[string][]*commonv1.SecretRef, index int, registry *ipgeo.Registry) (*proxyruntimev1.ProxyIPGeoProviderSettings, error) {
	if in == nil {
		return &proxyruntimev1.ProxyIPGeoProviderSettings{}, nil
	}
	id := strings.TrimSpace(in.GetProviderId())
	if id == "" {
		id = registry.DefaultProviderID(in.GetKind())
	}
	apiKeySecretRefs, err := ipGeoSecretRefsFromRequest(ctx, writer, in, id)
	if err != nil {
		return nil, err
	}
	if len(apiKeySecretRefs) == 0 && !in.GetClearApiKeys() && !in.GetAnonymous() {
		apiKeySecretRefs = current[ipGeoProviderSecretKey(in.GetKind(), id)]
	}
	weight := in.GetWeight()
	if weight == 0 {
		weight = settingscore.IPGeoProviderDefaultWeight(in.GetKind(), index, registry)
	}
	return &proxyruntimev1.ProxyIPGeoProviderSettings{
		ProviderId:       id,
		DisplayName:      strings.TrimSpace(in.GetDisplayName()),
		Weight:           weight,
		Kind:             in.GetKind(),
		Anonymous:        in.GetAnonymous(),
		ApiKeySecretRefs: apiKeySecretRefs,
	}, nil
}

func ipGeoSecretRefsFromRequest(ctx context.Context, writer secretref.Writer, in *proxyruntimev1.ProxyIPGeoProviderSettings, providerID string) ([]*commonv1.SecretRef, error) {
	if refs := settingscore.CleanIPGeoSecretRefs(in.GetApiKeySecretRefs()); len(refs) > 0 {
		return refs, nil
	}
	rawValues := appcore.CleanList(in.GetApiKeyValues())
	if len(rawValues) == 0 {
		return nil, nil
	}
	out := make([]*commonv1.SecretRef, 0, len(rawValues))
	for index, raw := range rawValues {
		secretID := secretref.StableID("proxy-runtime-ip-geo-api-key", fmt.Sprintf("%d", in.GetKind()), providerID, fmt.Sprintf("%d", index))
		saved, err := settingscore.WriteRuntimeSecret(ctx, writer, raw, secretID, settingscore.IPGeoAPIKeyPurpose)
		if err != nil {
			return nil, err
		}
		out = append(out, saved)
	}
	return settingscore.CleanIPGeoSecretRefs(out), nil
}
