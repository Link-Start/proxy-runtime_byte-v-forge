package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (a Application) Get(ctx context.Context) (*proxygatewayv1.GetProxyGatewaySettingsResponse, error) {
	repository, err := a.repositoryOrError()
	if err != nil {
		return nil, err
	}
	settings, err := repository.View(ctx)
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.GetProxyGatewaySettingsResponse{Settings: settings}, nil
}
