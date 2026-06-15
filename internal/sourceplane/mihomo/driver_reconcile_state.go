package mihomo

import "github.com/byte-v-forge/proxy-runtime/internal/sourceplane"

func (d *Driver) recordAppliedConfigProjection(configPath string, endpoint sourceplane.Endpoint, baseConfig renderedMihomoConfig, finalConfig renderedMihomoConfig) {
	d.signature = finalConfig.signature
	d.baseSig = baseConfig.signature
	d.configPath = configPath
	d.lastEndpoint = endpoint
	d.lastError = ""
}

func (d *Driver) recordConfigProjectionError(err error) error {
	if err == nil {
		return nil
	}
	d.lastError = err.Error()
	return err
}
