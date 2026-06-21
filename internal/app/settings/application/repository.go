package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type Repository interface {
	View(context.Context) (*proxygatewayv1.ProxyGatewaySettings, error)
	Load(context.Context) (*proxygatewayv1.ProxyGatewayPersistentSettings, error)
	Update(context.Context, *proxygatewayv1.UpdateProxyGatewaySettingsRequest) (*proxygatewayv1.ProxyGatewaySettings, error)
	UpdateDynamicIPProviders(context.Context, []*proxygatewayv1.ProxyDynamicIPProviderSettings) (*proxygatewayv1.ProxyGatewaySettings, error)
	UpdateEgressProfiles(context.Context, []*proxygatewayv1.EgressProfileSettings) (*proxygatewayv1.ProxyGatewaySettings, error)
	UpdateIngressRules(context.Context, []*proxygatewayv1.ProxyIngressRuleSettings) (*proxygatewayv1.ProxyGatewaySettings, error)
	UpdateInUserRules(context.Context, []*proxygatewayv1.EgressProfileSettings, []*proxygatewayv1.ProxyIngressRuleSettings) (*proxygatewayv1.ProxyGatewaySettings, error)
}
