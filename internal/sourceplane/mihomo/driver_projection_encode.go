package mihomo

import "encoding/json"

func encodeMihomoConfigProjection(configFile mihomoConfig) (renderedMihomoConfig, error) {
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return renderedMihomoConfig{}, err
	}
	return renderedMihomoConfig{data: data, signature: signature(data)}, nil
}
