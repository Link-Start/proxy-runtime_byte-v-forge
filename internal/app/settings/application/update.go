package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

func (a Application) UpdateRuntimeSettings(ctx context.Context, req *proxygatewayv1.UpdateProxyGatewaySettingsRequest) (*proxygatewayv1.UpdateProxyGatewaySettingsResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after update failed", func(repository Repository) (*proxygatewayv1.ProxyGatewaySettings, error) {
		return repository.Update(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyGatewaySettingsResponse{Settings: settings}, nil
}

func (a Application) UpdateEgressProfiles(ctx context.Context, req *proxygatewayv1.UpdateProxyEgressProfilesRequest) (*proxygatewayv1.UpdateProxyEgressProfilesResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after egress profile update failed", func(repository Repository) (*proxygatewayv1.ProxyGatewaySettings, error) {
		return repository.UpdateEgressProfiles(ctx, req.GetEgressProfiles())
	})
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyEgressProfilesResponse{Settings: settings}, nil
}

func (a Application) UpdateIngressRules(ctx context.Context, req *proxygatewayv1.UpdateProxyIngressRulesRequest) (*proxygatewayv1.UpdateProxyIngressRulesResponse, error) {
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after ingress rule update failed", func(repository Repository) (*proxygatewayv1.ProxyGatewaySettings, error) {
		return repository.UpdateIngressRules(ctx, req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyIngressRulesResponse{Settings: settings}, nil
}

func (a Application) UpdateInUserRules(ctx context.Context, req *proxygatewayv1.UpdateProxyGatewaySettingsRequest) (*proxygatewayv1.UpdateProxyGatewaySettingsResponse, error) {
	if err := a.validateProfiles(req.GetEgressProfiles()); err != nil {
		return nil, err
	}
	settings, err := a.updateWithConnectionCleanup(ctx, "load runtime settings after in-user rule update failed", func(repository Repository) (*proxygatewayv1.ProxyGatewaySettings, error) {
		return repository.UpdateInUserRules(ctx, req.GetEgressProfiles(), req.GetIngressRules())
	})
	if err != nil {
		return nil, err
	}
	return &proxygatewayv1.UpdateProxyGatewaySettingsResponse{Settings: settings}, nil
}
