package mihomo

import (
	"context"
	"path/filepath"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) applyBaseConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint) (baseConfigProjectionApplyResult, error) {
	decision := decideBaseConfigProjectionApply(baseConfigProjectionDecisionInput{
		running:          d.running,
		currentEndpoint:  d.lastEndpoint,
		nextEndpoint:     endpoint,
		currentSignature: d.baseSig,
		nextSignature:    config.signature,
	})
	switch decision.mode {
	case configProjectionApplyRestart:
		if err := d.restartConfigProjectionLocked(ctx, configPath, config, endpoint); err != nil {
			return baseConfigProjectionApplyResult{}, err
		}
	case configProjectionApplyReload:
		if err := d.reloadConfigDataLocked(ctx, configPath, config.data, endpoint); err != nil {
			return baseConfigProjectionApplyResult{}, err
		}
	}
	return decision.result(), nil
}

func (d *Driver) applyFinalConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint, baseApply baseConfigProjectionApplyResult) error {
	d.desiredSig = config.signature
	decision := decideFinalConfigProjectionApply(finalConfigProjectionDecisionInput{
		baseChanged:      baseApply.changed,
		currentSignature: d.signature,
		nextSignature:    config.signature,
	})
	if !decision.reloadRequired() {
		return nil
	}
	return d.reloadConfigDataLocked(ctx, configPath, config.data, endpoint)
}

func (d *Driver) restartConfigProjectionLocked(ctx context.Context, configPath string, config renderedMihomoConfig, endpoint sourceplane.Endpoint) error {
	if err := writeConfigData(configPath, config.data); err != nil {
		return err
	}
	d.stopLocked()
	if err := d.startLocked(ctx, filepath.Dir(configPath), configPath); err != nil {
		return err
	}
	return waitForReloadEndpoint(ctx, endpoint)
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
