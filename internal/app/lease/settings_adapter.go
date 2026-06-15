package lease

import (
	"context"
	"errors"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

var (
	ErrSettingsLoaderRequired           = errors.New("settings loader is required")
	ErrSettingsProviderGatewaysRequired = errors.New("settings provider gateways resolver is required")
	ErrSettingsRouteLineBindingRequired = errors.New("settings route line binding resolver is required")
)

type SettingsLoader[T any] func(context.Context) (T, error)

type SettingsProviderGatewaysResolver[T any] func(T, *proxyruntimev1.ProxyDynamicLease, string) ([]accountproxy.Gateway, error)

type SettingsRouteLineBindingResolver[T any] func(context.Context, T, string) (string, map[string]string, error)

type SettingsAdapter[T any] struct {
	Load                    SettingsLoader[T]
	ResolveProviderGateways SettingsProviderGatewaysResolver[T]
	ResolveLineBinding      SettingsRouteLineBindingResolver[T]
}

func (a SettingsAdapter[T]) ProviderGatewaysResolver(lease *proxyruntimev1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	return func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
		if a.Load == nil {
			return nil, ErrSettingsLoaderRequired
		}
		settings, err := a.Load(ctx)
		if err != nil {
			return nil, err
		}
		return a.ProviderGatewaysResolverForSettings(settings, lease)(ctx, providerID)
	}
}

func (a SettingsAdapter[T]) ProviderGatewaysResolverForSettings(settings T, lease *proxyruntimev1.ProxyDynamicLease) ProviderSessionGatewaysResolver {
	return func(ctx context.Context, providerID string) ([]accountproxy.Gateway, error) {
		_ = ctx
		if a.ResolveProviderGateways == nil {
			return nil, ErrSettingsProviderGatewaysRequired
		}
		return a.ResolveProviderGateways(settings, lease, providerID)
	}
}

func (a SettingsAdapter[T]) RouteLineBindingResolver(settings T) RouteLineBindingResolver {
	return func(ctx context.Context, accountID string) (string, map[string]string, error) {
		if a.ResolveLineBinding == nil {
			return "", nil, ErrSettingsRouteLineBindingRequired
		}
		return a.ResolveLineBinding(ctx, settings, accountID)
	}
}
