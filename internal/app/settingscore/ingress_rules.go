package settingscore

import (
	"fmt"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

func IngressRuleFromProto(in *proxyruntimev1.ProxyIngressRuleSettings, index int) *proxyruntimev1.ProxyIngressRuleSettings {
	if in == nil {
		return &proxyruntimev1.ProxyIngressRuleSettings{RuleId: fmt.Sprintf("ingress-%d", index+1)}
	}
	ruleID := appcore.RuntimeSafeID(in.GetRuleId())
	if ruleID == "" {
		ruleID = appcore.RuntimeSafeID(appcore.FirstNonEmpty(in.GetUsername(), in.GetDisplayName()))
	}
	if ruleID == "" {
		ruleID = fmt.Sprintf("ingress-%d", index+1)
	}
	return &proxyruntimev1.ProxyIngressRuleSettings{
		RuleId:        ruleID,
		DisplayName:   strings.TrimSpace(in.GetDisplayName()),
		Enabled:       in.GetEnabled(),
		Username:      strings.TrimSpace(in.GetUsername()),
		PasswordValue: in.GetPasswordValue(),
		ProfileId:     appcore.RuntimeSafeID(in.GetProfileId()),
	}
}
