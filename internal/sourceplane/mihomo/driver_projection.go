package mihomo

type renderedMihomoConfig struct {
	data      []byte
	signature string
}

func renderConfigProjection(options renderOptions) (renderedMihomoConfig, error) {
	configFile, err := projectMihomoConfigProjection(options)
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	if err := validateMihomoConfigProjection(configFile); err != nil {
		return renderedMihomoConfig{}, err
	}
	return encodeMihomoConfigProjection(configFile)
}
