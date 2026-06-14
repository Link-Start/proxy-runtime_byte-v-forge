package mihomo

import (
	"context"
	"path/filepath"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) applyBaseConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint) (bool, error) {
	restartRequired := !d.running || d.lastEndpoint != endpoint
	baseChanged := d.baseSig != config.signature
	if restartRequired {
		if err := writeConfigData(configPath, config.data); err != nil {
			return false, err
		}
		d.stopLocked()
		if err := d.startLocked(ctx, filepath.Dir(configPath), configPath); err != nil {
			return false, err
		}
		if err := waitForReloadEndpoint(ctx, endpoint); err != nil {
			return false, err
		}
		return true, nil
	}
	if baseChanged {
		if err := d.reloadConfigDataLocked(ctx, configPath, config.data, endpoint); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (d *Driver) applyFinalConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint, baseReloaded bool) error {
	d.desiredSig = config.signature
	if !baseReloaded && config.signature == d.signature {
		return nil
	}
	return d.reloadConfigDataLocked(ctx, configPath, config.data, endpoint)
}

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
