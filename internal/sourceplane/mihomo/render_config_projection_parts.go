package mihomo

type renderedConfigProjectionParts struct {
	proxy   renderedProxyProjection
	gateway renderedGatewayProjection
}

func renderConfigProjectionParts(opts renderOptions) (renderedConfigProjectionParts, error) {
	proxyProjection, err := renderProxyProjection(opts)
	if err != nil {
		return renderedConfigProjectionParts{}, err
	}
	gatewayProjection, err := renderGatewayProjection(opts, proxyProjection.profileGroupsBy)
	if err != nil {
		return renderedConfigProjectionParts{}, err
	}
	return renderedConfigProjectionParts{
		proxy:   proxyProjection,
		gateway: gatewayProjection,
	}, nil
}
