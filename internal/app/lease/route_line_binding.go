package lease

import (
	"context"
	"errors"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

var ErrRouteLineBindingResolverRequired = errors.New("route line binding resolver is required")

type RouteLineBindingResolver func(context.Context, string) (string, map[string]string, error)

type RouteLineBinding struct {
	Nodes       []provider.Node
	DialerProxy string
	Labels      map[string]string
}

func PrepareRouteLineBinding(ctx context.Context, nodes []provider.Node, accountID string, resolve RouteLineBindingResolver) (RouteLineBinding, error) {
	if resolve == nil {
		return RouteLineBinding{}, ErrRouteLineBindingResolverRequired
	}
	dialerProxy, labels, err := resolve(ctx, accountID)
	if err != nil {
		return RouteLineBinding{}, err
	}
	return RouteLineBinding{Nodes: ApplyNodeLabels(nodes, labels), DialerProxy: dialerProxy, Labels: labels}, nil
}
