package kernel

import (
	"fmt"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

func IngressRuleFromProto(in *proxygatewayv1.ProxyIngressRuleSettings, index int) *proxygatewayv1.ProxyIngressRuleSettings {
	if in == nil {
		return &proxygatewayv1.ProxyIngressRuleSettings{RuleId: fmt.Sprintf("ingress-%d", index+1)}
	}
	ruleID := appcore.RuntimeSafeID(in.GetRuleId())
	if ruleID == "" {
		ruleID = appcore.RuntimeSafeID(appcore.FirstNonEmpty(in.GetUsername(), in.GetDisplayName()))
	}
	if ruleID == "" {
		ruleID = fmt.Sprintf("ingress-%d", index+1)
	}
	return &proxygatewayv1.ProxyIngressRuleSettings{
		RuleId:        ruleID,
		DisplayName:   strings.TrimSpace(in.GetDisplayName()),
		Enabled:       in.GetEnabled(),
		Username:      strings.TrimSpace(in.GetUsername()),
		PasswordValue: in.GetPasswordValue(),
		ProfileId:     appcore.RuntimeSafeID(in.GetProfileId()),
	}
}
