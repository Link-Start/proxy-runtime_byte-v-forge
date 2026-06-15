package mihomo

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

type configProjectionApplyInput struct {
	configPath string
	config     renderedMihomoConfig
	endpoint   sourceplane.Endpoint
}

type finalConfigProjectionApplyInput struct {
	projection configProjectionApplyInput
	baseApply  baseConfigProjectionApplyResult
}

func (d *Driver) applyBaseConfigProjectionLocked(ctx context.Context, input configProjectionApplyInput) (baseConfigProjectionApplyResult, error) {
	decision := decideBaseConfigProjectionApply(baseConfigProjectionDecisionInput{
		running:          d.running,
		currentEndpoint:  d.lastEndpoint,
		nextEndpoint:     input.endpoint,
		currentSignature: d.baseSig,
		nextSignature:    input.config.signature,
	})
	if err := d.applyConfigProjectionModeLocked(ctx, decision.mode, input); err != nil {
		return baseConfigProjectionApplyResult{}, err
	}
	return decision.result(), nil
}

func (d *Driver) applyFinalConfigProjectionLocked(ctx context.Context, input finalConfigProjectionApplyInput) error {
	d.desiredSig = input.projection.config.signature
	decision := decideFinalConfigProjectionApply(finalConfigProjectionDecisionInput{
		baseChanged:      input.baseApply.changed,
		currentSignature: d.signature,
		nextSignature:    input.projection.config.signature,
	})
	return d.applyConfigProjectionModeLocked(ctx, decision.mode, input.projection)
}

func (d *Driver) applyConfigProjectionModeLocked(ctx context.Context, mode configProjectionApplyMode, input configProjectionApplyInput) error {
	switch mode {
	case configProjectionApplyNoop:
		return nil
	case configProjectionApplyRestart:
		return d.restartConfigProjectionLocked(ctx, input)
	case configProjectionApplyReload:
		return d.reloadConfigDataLocked(ctx, input.configPath, input.config.data, input.endpoint)
	default:
		return nil
	}
}
