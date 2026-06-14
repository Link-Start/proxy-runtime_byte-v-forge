package mihomo

import "encoding/json"

type renderedMihomoConfig struct {
	data      []byte
	signature string
}

func renderConfigProjection(options renderOptions) (renderedMihomoConfig, error) {
	configFile, err := renderConfig(options)
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	return renderedMihomoConfig{data: data, signature: signature(data)}, nil
}
