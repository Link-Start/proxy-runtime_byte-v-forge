package lease

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

func MarkCleanupPending(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool, finalStatus string) {
	if lease == nil {
		return
	}
	session := lease.GetSession()
	if session == nil {
		session = &proxyruntimev1.ProxySession{}
		lease.Session = session
	}
	if session.Labels == nil {
		session.Labels = map[string]string{}
	}
	if routePending {
		session.Labels[RouteCleanupPendingLabel] = "true"
	}
	if providerPending {
		session.Labels[ProviderCleanupPendingLabel] = "true"
	}
	if strings.TrimSpace(finalStatus) != "" {
		session.Labels[CleanupFinalStatusLabel] = strings.TrimSpace(finalStatus)
	}
}

func ClearCleanupPending(lease *proxyruntimev1.ProxyDynamicLease, routePending bool, providerPending bool) {
	if lease == nil || lease.GetSession() == nil || lease.GetSession().Labels == nil {
		return
	}
	if routePending {
		delete(lease.GetSession().Labels, RouteCleanupPendingLabel)
	}
	if providerPending {
		delete(lease.GetSession().Labels, ProviderCleanupPendingLabel)
	}
}
