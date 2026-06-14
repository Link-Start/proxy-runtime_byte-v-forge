package mihomo

func renderConfig(opts renderOptions) (mihomoConfig, error) {
	sections, err := renderConfigSections(opts)
	if err != nil {
		return mihomoConfig{}, err
	}
	return assembleRenderedConfig(opts, sections), nil
}
