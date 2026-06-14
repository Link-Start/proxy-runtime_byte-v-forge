package lease

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

type DynamicListenerInput struct {
	ID                  string
	Addr                string
	Protocol            string
	Route               string
	AccountID           string
	LeaseID             string
	DefaultUsername     string
	FallbackPassword    string
	IngressRules        []*proxyruntimev1.ProxyIngressRuleSettings
	PlaygroundAccountID string
	PlaygroundRuleID    string
	PlaygroundUsername  string
}

func NewDynamicListener(input DynamicListenerInput) (Listener, error) {
	return NewListener(ListenerInput{
		ID:        input.ID,
		Addr:      input.Addr,
		Protocol:  input.Protocol,
		Route:     input.Route,
		Username:  dynamicListenerUsername(input),
		Password:  dynamicListenerPassword(input),
		AccountID: input.AccountID,
		LeaseID:   input.LeaseID,
	})
}

func dynamicListenerUsername(input DynamicListenerInput) string {
	if input.AccountID == input.PlaygroundAccountID {
		if rule := PlaygroundIngressRule(input.IngressRules, input.PlaygroundRuleID, input.PlaygroundUsername); rule != nil {
			return rule.GetUsername()
		}
	}
	return input.DefaultUsername
}

func dynamicListenerPassword(input DynamicListenerInput) string {
	password := ListenerPassword(input.IngressRules, input.AccountID, input.FallbackPassword)
	if input.AccountID == input.PlaygroundAccountID {
		if rule := PlaygroundIngressRule(input.IngressRules, input.PlaygroundRuleID, input.PlaygroundUsername); rule != nil {
			return firstNonEmpty(rule.GetPasswordValue(), password)
		}
	}
	return password
}
