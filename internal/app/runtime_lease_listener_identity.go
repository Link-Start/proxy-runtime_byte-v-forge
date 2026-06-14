package app

func proxyRouteUsername(accountID string) string {
	username := runtimeSafeID(accountID)
	if username == "" {
		username = shortHash(accountID)
	}
	return "acct-" + username
}
