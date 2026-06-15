package mihomo

import "github.com/byte-v-forge/proxy-runtime/internal/sourceplane"

type configProjectionApplyMode uint8

const (
	configProjectionApplyNoop configProjectionApplyMode = iota
	configProjectionApplyReload
	configProjectionApplyRestart
)

type baseConfigProjectionDecision struct {
	mode configProjectionApplyMode
}

type finalConfigProjectionDecision struct {
	mode configProjectionApplyMode
}

type baseConfigProjectionApplyResult struct {
	changed bool
}

type baseConfigProjectionDecisionInput struct {
	running          bool
	currentEndpoint  sourceplane.Endpoint
	nextEndpoint     sourceplane.Endpoint
	currentSignature string
	nextSignature    string
}

type finalConfigProjectionDecisionInput struct {
	baseChanged      bool
	currentSignature string
	nextSignature    string
}

func decideBaseConfigProjectionApply(input baseConfigProjectionDecisionInput) baseConfigProjectionDecision {
	if !input.running || input.currentEndpoint != input.nextEndpoint {
		return baseConfigProjectionDecision{mode: configProjectionApplyRestart}
	}
	if input.currentSignature != input.nextSignature {
		return baseConfigProjectionDecision{mode: configProjectionApplyReload}
	}
	return baseConfigProjectionDecision{mode: configProjectionApplyNoop}
}

func decideFinalConfigProjectionApply(input finalConfigProjectionDecisionInput) finalConfigProjectionDecision {
	if !input.baseChanged && input.currentSignature == input.nextSignature {
		return finalConfigProjectionDecision{mode: configProjectionApplyNoop}
	}
	return finalConfigProjectionDecision{mode: configProjectionApplyReload}
}

func (d baseConfigProjectionDecision) changed() bool {
	return d.mode != configProjectionApplyNoop
}

func (d baseConfigProjectionDecision) result() baseConfigProjectionApplyResult {
	return baseConfigProjectionApplyResult{changed: d.changed()}
}
