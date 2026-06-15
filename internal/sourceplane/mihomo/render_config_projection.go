package mihomo

type renderedConfigSections struct {
	listener  mihomoListener
	proxies   []map[string]any
	providers map[string]mihomoProvider
	groups    []mihomoGroup
	rules     []string
}

func renderConfigSections(opts renderOptions) (renderedConfigSections, error) {
	parts, err := renderConfigProjectionParts(opts)
	if err != nil {
		return renderedConfigSections{}, err
	}
	return assembleRenderedConfigSections(opts, parts), nil
}
