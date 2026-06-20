package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	settingsdomain "github.com/byte-v-forge/proxy-runtime/internal/app/settings/domain"
)

func (s *runtimeSettingsStore) View(ctx context.Context) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	settings, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return settingsdomain.RuntimeSettingsView(settings), nil
}
