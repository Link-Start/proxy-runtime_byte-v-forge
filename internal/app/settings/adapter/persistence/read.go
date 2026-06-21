package persistence

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	settingsdomain "github.com/byte-v-forge/proxy-gateway/internal/app/settings/domain"
)

func (s *Store) View(ctx context.Context) (*proxygatewayv1.ProxyGatewaySettings, error) {
	settings, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return settingsdomain.RuntimeSettingsView(settings), nil
}
