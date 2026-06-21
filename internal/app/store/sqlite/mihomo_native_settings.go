package sqlite

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/app/mihomonative"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) LoadMihomoNativeSettings(ctx context.Context) (*proxygatewayv1.ProxyGatewayMihomoNativeConfig, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, store.MihomoNativeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return mihomonative.NormalizeSettings(nil), nil
	}
	return mihomonative.DecodeMihomoNativeSettings(raw)
}

func (s *Store) SaveMihomoNativeSettings(ctx context.Context, settings *proxygatewayv1.ProxyGatewayMihomoNativeConfig) error {
	data, err := protojsoncodec.Marshal(mihomonative.NormalizeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, store.MihomoNativeSettingsKey, data)
}
