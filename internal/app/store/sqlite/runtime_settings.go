package sqlite

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"

	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
)

func (s *Store) LoadRuntimeSettings(ctx context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error) {
	raw, found, err := s.loadRuntimeSettingJSON(ctx, store.RuntimeSettingsKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return kernel.NormalizeRuntimeSettings(nil), nil
	}
	return kernel.DecodeRuntimeSettings(raw)
}

func (s *Store) SaveRuntimeSettings(ctx context.Context, settings *proxygatewayv1.ProxyGatewayPersistentSettings) error {
	data, err := protojsoncodec.Marshal(kernel.NormalizeRuntimeSettings(settings))
	if err != nil {
		return err
	}
	return s.saveRuntimeSettingJSON(ctx, store.RuntimeSettingsKey, data)
}
