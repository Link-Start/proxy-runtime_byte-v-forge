package application

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func (a Application) Get(ctx context.Context) (*proxyruntimev1.GetProxyRuntimeSettingsResponse, error) {
	repository, err := a.repositoryOrError()
	if err != nil {
		return nil, err
	}
	settings, err := repository.View(ctx)
	if err != nil {
		return nil, err
	}
	return &proxyruntimev1.GetProxyRuntimeSettingsResponse{Settings: settings}, nil
}
