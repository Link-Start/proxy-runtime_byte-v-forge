package app

func leaseListenerUsername(settings *runtimeSettingsFile, accountID string, leaseID string) string {
	username := proxyRouteUsername(leaseID)
	if accountID == playgroundProfileID {
		if rule := playgroundIngressRule(settings); rule != nil {
			return rule.GetUsername()
		}
	}
	return username
}

func leaseListenerPasswordValue(settings *runtimeSettingsFile, accountID string, fallback string) string {
	password := leaseListenerPassword(settings, accountID, fallback)
	if accountID == playgroundProfileID {
		if rule := playgroundIngressRule(settings); rule != nil {
			return firstNonEmpty(rule.GetPasswordValue(), password)
		}
	}
	return password
}
