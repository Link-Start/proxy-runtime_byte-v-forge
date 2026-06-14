package app

import leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"

func leaseListenerPassword(settings *runtimeSettingsFile, profileID string, fallback string) string {
	return leaseapp.ListenerPassword(settings.GetIngressRules(), profileID, fallback)
}
