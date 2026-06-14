package mihomo

func renderBaseProxyConfigs(opts renderOptions) ([]map[string]any, error) {
	fixedConfigs := cloneNativeProxies(opts.NativeConfig.Proxies)
	poolConfigs, _, err := renderProviderNodes("provider-pool", opts.BasePool)
	if err != nil {
		return nil, err
	}
	fixedConfigs = append(fixedConfigs, poolConfigs...)
	sessionConfigs, err := renderSessionRoutes(opts.SessionRoutes)
	if err != nil {
		return nil, err
	}
	return append(fixedConfigs, sessionConfigs...), nil
}
