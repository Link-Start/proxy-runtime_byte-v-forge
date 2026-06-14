package mihomo

type renderedGatewayProjection struct {
	listener mihomoListener
	groups   []mihomoGroup
	rules    []string
}

func renderGatewayProjection(opts renderOptions, profileGroupsByID map[string]string) (renderedGatewayProjection, error) {
	gateway, err := renderGateway(opts.Endpoint, opts.ProxyUsers, opts.SessionRoutes, profileGroupsByID)
	if err != nil {
		return renderedGatewayProjection{}, err
	}
	return renderedGatewayProjection{listener: gateway.listener, groups: gateway.groups, rules: gateway.rules}, nil
}
