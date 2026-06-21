package domain

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/app/kernel"
)

const (
	playgroundDynamicRuleID    = "in-user-playground-dynamic"
	playgroundDynamicProfileID = "playground-dynamic-egress"
	playgroundDynamicUsername  = "playground-dynamic"
)

func ensurePlaygroundInUserRules(settings *proxygatewayv1.ProxyGatewayPersistentSettings) bool {
	if settings == nil || !hasPlaygroundRule(settings.GetIngressRules()) || playgroundAlreadySingle(settings) {
		return false
	}
	sourceRule := playgroundSourceIngressRule(settings.GetIngressRules())
	sourceProfile := playgroundSourceProfile(settings, sourceRule.GetProfileId())
	settings.EgressProfiles = replacePlaygroundProfiles(settings.GetEgressProfiles(), playgroundSingleProfile(sourceProfile))
	settings.IngressRules = replacePlaygroundRules(settings.GetIngressRules(), playgroundSingleRule(sourceRule))
	return true
}

func playgroundAlreadySingle(settings *proxygatewayv1.ProxyGatewayPersistentSettings) bool {
	found := false
	for _, rule := range settings.GetIngressRules() {
		if rule.GetRuleId() == kernel.PlaygroundRuleID && strings.TrimSpace(rule.GetUsername()) == kernel.PlaygroundUsername && rule.GetProfileId() == kernel.PlaygroundProfileID {
			found = true
			continue
		}
		if isPlaygroundRule(rule) {
			return false
		}
	}
	if !found {
		return false
	}
	return playgroundProfileByID(settings, playgroundDynamicProfileID) == nil
}

func hasPlaygroundRule(rules []*proxygatewayv1.ProxyIngressRuleSettings) bool {
	for _, rule := range rules {
		if isPlaygroundRule(rule) {
			return true
		}
	}
	return false
}

func isPlaygroundRule(rule *proxygatewayv1.ProxyIngressRuleSettings) bool {
	username := strings.TrimSpace(rule.GetUsername())
	return rule.GetRuleId() == kernel.PlaygroundRuleID || rule.GetRuleId() == playgroundDynamicRuleID || username == kernel.PlaygroundUsername || strings.HasPrefix(username, kernel.PlaygroundUsername+"-session-") || username == playgroundDynamicUsername || strings.HasPrefix(username, playgroundDynamicUsername+"-session-")
}

func playgroundSourceIngressRule(rules []*proxygatewayv1.ProxyIngressRuleSettings) *proxygatewayv1.ProxyIngressRuleSettings {
	for _, rule := range rules {
		if rule.GetRuleId() == kernel.PlaygroundRuleID || strings.HasPrefix(strings.TrimSpace(rule.GetUsername()), kernel.PlaygroundUsername) {
			return rule
		}
	}
	for _, rule := range rules {
		if isPlaygroundRule(rule) {
			return rule
		}
	}
	return nil
}

func playgroundSourceProfile(settings *proxygatewayv1.ProxyGatewayPersistentSettings, profileID string) *proxygatewayv1.EgressProfileSettings {
	if profile := playgroundProfileByID(settings, playgroundDynamicProfileID); profile != nil && profile.GetExit().GetKind() == proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return profile
	}
	if profile := playgroundProfileByID(settings, profileID); profile != nil {
		return profile
	}
	return playgroundProfileByID(settings, kernel.PlaygroundProfileID)
}

func playgroundSingleRule(source *proxygatewayv1.ProxyIngressRuleSettings) *proxygatewayv1.ProxyIngressRuleSettings {
	return &proxygatewayv1.ProxyIngressRuleSettings{RuleId: kernel.PlaygroundRuleID, DisplayName: "PlayGround", Enabled: true, Username: kernel.PlaygroundUsername, PasswordValue: source.GetPasswordValue(), ProfileId: kernel.PlaygroundProfileID}
}

func playgroundSingleProfile(source *proxygatewayv1.EgressProfileSettings) *proxygatewayv1.EgressProfileSettings {
	if source != nil {
		profile := cloneEgressProfile(source)
		profile.ProfileId = kernel.PlaygroundProfileID
		profile.DisplayName = "PlayGround"
		profile.Enabled = true
		return profile
	}
	return playgroundDirectProfile()
}

func playgroundDirectProfile() *proxygatewayv1.EgressProfileSettings {
	return &proxygatewayv1.EgressProfileSettings{ProfileId: kernel.PlaygroundProfileID, DisplayName: "PlayGround", Enabled: true, Line: &proxygatewayv1.EgressProfileLineSettings{Kind: proxygatewayv1.EgressProfileLineKind_EGRESS_PROFILE_LINE_KIND_DIRECT}, Exit: &proxygatewayv1.EgressProfileExitSettings{Kind: proxygatewayv1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DIRECT}}
}

func playgroundProfileByID(settings *proxygatewayv1.ProxyGatewayPersistentSettings, profileID string) *proxygatewayv1.EgressProfileSettings {
	profileID = appcore.RuntimeSafeID(profileID)
	for _, profile := range settings.GetEgressProfiles() {
		if profile.GetProfileId() == profileID {
			return profile
		}
	}
	return nil
}

func replacePlaygroundProfiles(profiles []*proxygatewayv1.EgressProfileSettings, profile *proxygatewayv1.EgressProfileSettings) []*proxygatewayv1.EgressProfileSettings {
	out := make([]*proxygatewayv1.EgressProfileSettings, 0, len(profiles)+1)
	inserted := false
	for _, current := range profiles {
		if current.GetProfileId() == playgroundDynamicProfileID {
			continue
		}
		if current.GetProfileId() == kernel.PlaygroundProfileID {
			if !inserted {
				out = append(out, profile)
				inserted = true
			}
			continue
		}
		out = append(out, current)
	}
	if !inserted {
		out = append(out, profile)
	}
	return out
}

func replacePlaygroundRules(rules []*proxygatewayv1.ProxyIngressRuleSettings, rule *proxygatewayv1.ProxyIngressRuleSettings) []*proxygatewayv1.ProxyIngressRuleSettings {
	out := make([]*proxygatewayv1.ProxyIngressRuleSettings, 0, len(rules)+1)
	inserted := false
	for _, current := range rules {
		if isPlaygroundRule(current) {
			if !inserted {
				out = append(out, rule)
				inserted = true
			}
			continue
		}
		out = append(out, current)
	}
	if !inserted {
		out = append(out, rule)
	}
	return out
}
