package app

import "github.com/byte-v-forge/proxy-runtime/internal/app/appcore"

func proxyRouteUsername(accountID string) string {
	username := appcore.RuntimeSafeID(accountID)
	if username == "" {
		username = appcore.ShortHash(accountID)
	}
	return "acct-" + username
}
